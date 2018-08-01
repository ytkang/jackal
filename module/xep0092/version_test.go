/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package xep0092

import (
	"testing"

	"github.com/ortuman/jackal/host"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/storage"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/version"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
)

func TestXEP0092(t *testing.T) {
	host.Initialize([]host.Config{{Name: "jackal.im"}})
	router.Initialize(&router.Config{})
	storage.Initialize(&storage.Config{Type: storage.Memory})
	defer func() {
		storage.Shutdown()
		router.Shutdown()
		host.Shutdown()
	}()
	srvJID, _ := jid.New("", "jackal.im", "", true)
	j, _ := jid.New("ortuman", "jackal.im", "balcony", true)

	stm := stream.NewMockC2S("abcd", j)
	router.Bind(stm)

	cfg := Config{}
	x := New(&cfg, nil)

	// test MatchesIQ
	iq := xmpp.NewIQType(uuid.New(), xmpp.GetType)
	iq.SetFrom(j.String())
	iq.SetTo(j.String())

	qVer := xmpp.NewElementNamespace("query", versionNamespace)

	iq.AppendElement(xmpp.NewElementNamespace("query", "jabber:client"))
	require.False(t, x.MatchesIQ(iq))
	iq.ClearElements()
	iq.AppendElement(qVer)
	require.False(t, x.MatchesIQ(iq))
	iq.SetTo(srvJID.String())
	require.True(t, x.MatchesIQ(iq))

	qVer.AppendElement(xmpp.NewElementName("version"))
	x.ProcessIQ(iq)
	elem := stm.FetchElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	// get version
	qVer.ClearElements()
	x.ProcessIQ(iq)
	elem = stm.FetchElement()
	ver := elem.Elements().ChildNamespace("query", versionNamespace)
	require.Equal(t, "jackal", ver.Elements().Child("name").Text())
	require.Equal(t, version.ApplicationVersion.String(), ver.Elements().Child("version").Text())
	require.Nil(t, ver.Elements().Child("os"))

	// show OS
	cfg.ShowOS = true

	x = New(&cfg, nil)
	x.ProcessIQ(iq)
	elem = stm.FetchElement()
	ver = elem.Elements().ChildNamespace("query", versionNamespace)
	require.Equal(t, osString, ver.Elements().Child("os").Text())
}
