/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package component

import "github.com/ortuman/jackal/xmpp"

type Component interface {
	Name() string
	ProcessStanza(stanza xmpp.Stanza)
}

func Register(comp Component) error {
	return nil
}

func Unregister(comp Component) error {
	return nil
}
