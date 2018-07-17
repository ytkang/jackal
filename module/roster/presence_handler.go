/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package roster

import (
	"sync"

	"github.com/ortuman/jackal/host"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/model/rostermodel"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/storage"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

var onlineJIDs sync.Map

// OnlinePresencesMatchingJID returns current online presences matching a given JID.
func OnlinePresencesMatchingJID(j *jid.JID) []*xmpp.Presence {
	var ret []*xmpp.Presence
	onlineJIDs.Range(func(_, value interface{}) bool {
		switch presence := value.(type) {
		case *xmpp.Presence:
			if onlineJIDMatchesJID(presence.FromJID(), j) {
				ret = append(ret, presence)
			}
		}
		return true
	})
	return ret
}

func onlineJIDMatchesJID(onlineJID, j *jid.JID) bool {
	if j.IsFullWithUser() {
		return onlineJID.Matches(j, jid.MatchesNode|jid.MatchesDomain|jid.MatchesResource)
	} else if j.IsFullWithServer() {
		return onlineJID.Matches(j, jid.MatchesDomain|jid.MatchesResource)
	} else if j.IsBare() {
		return onlineJID.Matches(j, jid.MatchesNode|jid.MatchesDomain)
	}
	return onlineJID.Matches(j, jid.MatchesDomain)
}

// PresenceHandler represents a roster presence handler.
type PresenceHandler struct {
	cfg *Config
}

// NewPresenceHandler returns a new presence handler instance.
func NewPresenceHandler(cfg *Config) *PresenceHandler {
	return &PresenceHandler{cfg: cfg}
}

// ProcessPresence processes an incoming presence stanza.
func (ph *PresenceHandler) ProcessPresence(presence *xmpp.Presence) error {
	switch presence.Type() {
	case xmpp.SubscribeType:
		return ph.processSubscribe(presence)
	case xmpp.SubscribedType:
		return ph.processSubscribed(presence)
	case xmpp.UnsubscribeType:
		return ph.processUnsubscribe(presence)
	case xmpp.UnsubscribedType:
		return ph.processUnsubscribed(presence)
	case xmpp.ProbeType:
		return ph.processProbePresence(presence)
	case xmpp.AvailableType, xmpp.UnavailableType:
		return ph.processAvailablePresence(presence)
	}
	return nil
}

func (ph *PresenceHandler) processSubscribe(presence *xmpp.Presence) error {
	userJID := presence.FromJID().ToBareJID()
	contactJID := presence.ToJID().ToBareJID()

	log.Infof("processing 'subscribe' - contact: %s (%s)", contactJID, userJID)

	if host.IsLocalHost(userJID.Domain()) {
		usrRi, err := storage.Instance().FetchRosterItem(userJID.Node(), contactJID.String())
		if err != nil {
			return err
		}
		if usrRi != nil {
			switch usrRi.Subscription {
			case rostermodel.SubscriptionTo, rostermodel.SubscriptionBoth:
				return nil // already subscribed...
			default:
				if !usrRi.Ask {
					usrRi.Ask = true
				} else {
					return nil // notification already sent...
				}
			}
		} else {
			// create roster item if not previously created
			usrRi = &rostermodel.Item{
				Username:     userJID.Node(),
				JID:          contactJID.String(),
				Subscription: rostermodel.SubscriptionNone,
				Ask:          true,
			}
		}
		if insertItem(usrRi, userJID, ph.cfg.Versioning); err != nil {
			return err
		}
	}
	// stamp the presence stanza of type "subscribe" with the user's bare JID as the 'from' address
	p := xmpp.NewPresence(userJID, contactJID, xmpp.SubscribeType)
	p.AppendElements(presence.Elements().All())

	if host.IsLocalHost(contactJID.Domain()) {
		// archive roster approval notification
		if err := insertOrUpdateNotification(contactJID.Node(), userJID, p); err != nil {
			return err
		}
	}
	router.Route(p)
	return nil
}

func (ph *PresenceHandler) processSubscribed(presence *xmpp.Presence) error {
	userJID := presence.ToJID().ToBareJID()
	contactJID := presence.FromJID().ToBareJID()

	log.Infof("processing 'subscribed' - user: %s (%s)", userJID, contactJID)

	if host.IsLocalHost(contactJID.Domain()) {
		_, err := deleteNotification(contactJID.Node(), userJID)
		if err != nil {
			return err
		}
		cntRi, err := storage.Instance().FetchRosterItem(contactJID.Node(), userJID.String())
		if err != nil {
			return err
		}
		if cntRi != nil {
			switch cntRi.Subscription {
			case rostermodel.SubscriptionTo:
				cntRi.Subscription = rostermodel.SubscriptionBoth
			case rostermodel.SubscriptionNone:
				cntRi.Subscription = rostermodel.SubscriptionFrom
			}
		} else {
			// create roster item if not previously created
			cntRi = &rostermodel.Item{
				Username:     contactJID.Node(),
				JID:          userJID.String(),
				Subscription: rostermodel.SubscriptionFrom,
				Ask:          false,
			}
		}
		if insertItem(cntRi, contactJID, ph.cfg.Versioning); err != nil {
			return err
		}
	}
	// stamp the presence stanza of type "subscribed" with the contact's bare JID as the 'from' address
	p := xmpp.NewPresence(contactJID, userJID, xmpp.SubscribedType)
	p.AppendElements(presence.Elements().All())

	if host.IsLocalHost(userJID.Domain()) {
		usrRi, err := storage.Instance().FetchRosterItem(userJID.Node(), contactJID.String())
		if err != nil {
			return err
		}
		if usrRi != nil {
			switch usrRi.Subscription {
			case rostermodel.SubscriptionFrom:
				usrRi.Subscription = rostermodel.SubscriptionBoth
			case rostermodel.SubscriptionNone:
				usrRi.Subscription = rostermodel.SubscriptionTo
			default:
				return nil
			}
			usrRi.Ask = false
			if insertItem(usrRi, userJID, ph.cfg.Versioning); err != nil {
				return err
			}
		}
	}
	router.Route(p)
	routePresencesFrom(contactJID, userJID, xmpp.AvailableType)
	return nil
}

func (ph *PresenceHandler) processUnsubscribe(presence *xmpp.Presence) error {
	userJID := presence.FromJID().ToBareJID()
	contactJID := presence.ToJID().ToBareJID()

	log.Infof("processing 'unsubscribe' - contact: %s (%s)", contactJID, userJID)

	var usrSub string
	if host.IsLocalHost(userJID.Domain()) {
		usrRi, err := storage.Instance().FetchRosterItem(userJID.Node(), contactJID.String())
		if err != nil {
			return err
		}
		usrSub = rostermodel.SubscriptionNone
		if usrRi != nil {
			usrSub = usrRi.Subscription
			switch usrSub {
			case rostermodel.SubscriptionBoth:
				usrRi.Subscription = rostermodel.SubscriptionFrom
			default:
				usrRi.Subscription = rostermodel.SubscriptionNone
			}
			if insertItem(usrRi, userJID, ph.cfg.Versioning); err != nil {
				return err
			}
		}
	}
	// stamp the presence stanza of type "unsubscribe" with the users's bare JID as the 'from' address
	p := xmpp.NewPresence(userJID, contactJID, xmpp.UnsubscribeType)
	p.AppendElements(presence.Elements().All())

	if host.IsLocalHost(contactJID.Domain()) {
		cntRi, err := storage.Instance().FetchRosterItem(contactJID.Node(), userJID.String())
		if err != nil {
			return err
		}
		if cntRi != nil {
			switch cntRi.Subscription {
			case rostermodel.SubscriptionBoth:
				cntRi.Subscription = rostermodel.SubscriptionTo
			default:
				cntRi.Subscription = rostermodel.SubscriptionNone
			}
			if insertItem(cntRi, contactJID, ph.cfg.Versioning); err != nil {
				return err
			}
		}
	}
	router.Route(p)

	if usrSub == rostermodel.SubscriptionTo || usrSub == rostermodel.SubscriptionBoth {
		routePresencesFrom(contactJID, userJID, xmpp.UnavailableType)
	}
	return nil
}

func (ph *PresenceHandler) processUnsubscribed(presence *xmpp.Presence) error {
	userJID := presence.ToJID().ToBareJID()
	contactJID := presence.FromJID().ToBareJID()

	log.Infof("processing 'unsubscribed' - user: %s (%s)", userJID, contactJID)

	var cntSub string
	if host.IsLocalHost(contactJID.Domain()) {
		deleted, err := deleteNotification(contactJID.Node(), userJID)
		if err != nil {
			return err
		}
		// do not change subscription state if cancelling a subscription request
		if deleted {
			goto routePresence
		}
		cntRi, err := storage.Instance().FetchRosterItem(contactJID.Node(), userJID.String())
		if err != nil {
			return err
		}
		cntSub = rostermodel.SubscriptionNone
		if cntRi != nil {
			cntSub = cntRi.Subscription
			switch cntSub {
			case rostermodel.SubscriptionBoth:
				cntRi.Subscription = rostermodel.SubscriptionTo
			default:
				cntRi.Subscription = rostermodel.SubscriptionNone
			}
			if insertItem(cntRi, contactJID, ph.cfg.Versioning); err != nil {
				return err
			}
		}
	}
routePresence:
	// stamp the presence stanza of type "unsubscribed" with the contact's bare JID as the 'from' address
	p := xmpp.NewPresence(contactJID, userJID, xmpp.UnsubscribedType)
	p.AppendElements(presence.Elements().All())

	if host.IsLocalHost(userJID.Domain()) {
		usrRi, err := storage.Instance().FetchRosterItem(userJID.Node(), contactJID.String())
		if err != nil {
			return err
		}
		if usrRi != nil {
			if !usrRi.Ask { // pending out...
				switch usrRi.Subscription {
				case rostermodel.SubscriptionBoth:
					usrRi.Subscription = rostermodel.SubscriptionFrom
				default:
					usrRi.Subscription = rostermodel.SubscriptionNone
				}
			}
			usrRi.Ask = false
			if insertItem(usrRi, userJID, ph.cfg.Versioning); err != nil {
				return err
			}
		}
	}
	router.Route(p)

	if cntSub == rostermodel.SubscriptionFrom || cntSub == rostermodel.SubscriptionBoth {
		routePresencesFrom(contactJID, userJID, xmpp.UnavailableType)
	}
	return nil
}

func (ph *PresenceHandler) processProbePresence(presence *xmpp.Presence) error {
	userJID := presence.ToJID().ToBareJID()
	contactJID := presence.FromJID().ToBareJID()

	log.Infof("processing 'probe' - user: %s (%s)", userJID, contactJID)

	ri, err := storage.Instance().FetchRosterItem(userJID.Node(), contactJID.String())
	if err != nil {
		return err
	}
	usr, err := storage.Instance().FetchUser(userJID.Node())
	if err != nil {
		return err
	}
	if usr == nil || ri == nil || (ri.Subscription != rostermodel.SubscriptionBoth && ri.Subscription != rostermodel.SubscriptionFrom) {
		router.Route(xmpp.NewPresence(userJID, contactJID, xmpp.UnsubscribedType))
		return nil
	}
	if usr.LastPresence != nil {
		p := xmpp.NewPresence(usr.LastPresence.FromJID(), contactJID, usr.LastPresence.Type())
		p.AppendElements(usr.LastPresence.Elements().All())
		router.Route(p)
	}
	return nil
}

func (ph *PresenceHandler) processAvailablePresence(presence *xmpp.Presence) error {
	fromJID := presence.FromJID()

	userJID := fromJID.ToBareJID()
	contactJID := presence.ToJID().ToBareJID()

	replyOnBehalf := host.IsLocalHost(userJID.Domain()) && userJID.Matches(contactJID, jid.MatchesBare)

	// keep track of available presences
	if presence.IsAvailable() {
		log.Infof("processing 'available' - user: %s", fromJID)
		if _, loaded := onlineJIDs.LoadOrStore(fromJID.String(), presence); !loaded {
			if replyOnBehalf {
				if err := ph.deliverRosterPresences(userJID); err != nil {
					return err
				}
			}
		}
	} else {
		log.Infof("processing 'unavailable' - user: %s", fromJID)
		onlineJIDs.Delete(fromJID.String())
	}
	if replyOnBehalf {
		return ph.broadcastPresence(presence)
	}
	return router.Route(presence)
}

func (ph *PresenceHandler) deliverRosterPresences(userJID *jid.JID) error {
	// first, deliver pending approval notifications...
	rns, err := storage.Instance().FetchRosterNotifications(userJID.Node())
	if err != nil {
		return err
	}
	for _, rn := range rns {
		fromJID, _ := jid.NewWithString(rn.JID, true)
		p := xmpp.NewPresence(fromJID, userJID, xmpp.SubscribeType)
		p.AppendElements(rn.Presence.Elements().All())
		router.Route(p)
	}

	// deliver roster online presences
	items, _, err := storage.Instance().FetchRosterItems(userJID.Node())
	if err != nil {
		return err
	}
	for _, item := range items {
		switch item.Subscription {
		case rostermodel.SubscriptionTo, rostermodel.SubscriptionBoth:
			contactJID := item.ContactJID()
			if !host.IsLocalHost(contactJID.Domain()) {
				router.Route(xmpp.NewPresence(userJID, contactJID, xmpp.ProbeType))
				continue
			}
			routePresencesFrom(contactJID, userJID, xmpp.AvailableType)
		}
	}
	return nil
}

func (ph *PresenceHandler) broadcastPresence(presence *xmpp.Presence) error {
	fromJID := presence.FromJID()
	itms, _, err := storage.Instance().FetchRosterItems(fromJID.Node())
	if err != nil {
		return err
	}
	for _, itm := range itms {
		switch itm.Subscription {
		case rostermodel.SubscriptionFrom, rostermodel.SubscriptionBoth:
			p := xmpp.NewPresence(fromJID, itm.ContactJID(), presence.Type())
			p.AppendElements(presence.Elements().All())
			router.Route(p)
		}
	}

	// update last received presence
	if usr, err := storage.Instance().FetchUser(fromJID.Node()); err != nil {
		return err
	} else if usr != nil {
		usr.LastPresence = presence
		return storage.Instance().InsertOrUpdateUser(usr)
	}
	return nil
}
