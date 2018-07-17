/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package xep0199

import (
	"testing"

	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
)

func TestXEP0199_Matching(t *testing.T) {
	t.Parallel()
	j, _ := jid.New("ortuman", "jackal.im", "balcony", true)

	x := New(&Config{}, nil)

	// test MatchesIQ
	iqID := uuid.New()
	iq := xmpp.NewIQType(iqID, xmpp.GetType)
	iq.SetFromJID(j)

	ping := xmpp.NewElementNamespace("ping", pingNamespace)
	iq.AppendElement(ping)

	require.True(t, x.MatchesIQ(iq))
}

func TestXEP0199_ReceivePing(t *testing.T) {
	t.Parallel()
	j1, _ := jid.New("ortuman", "jackal.im", "balcony", true)
	j2, _ := jid.New("juliet", "jackal.im", "garden", true)

	stm := stream.NewMockC2S("abcd", j1)
	defer stm.Disconnect(nil)

	stm.SetUsername("ortuman")

	x := New(&Config{}, stm)

	iqID := uuid.New()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j2)
	iq.SetToJID(j2)

	x.ProcessIQ(iq)
	elem := stm.FetchElement()
	require.Equal(t, xmpp.ErrForbidden.Error(), elem.Error().Elements().All()[0].Name())

	iq.SetToJID(j1)
	x.ProcessIQ(iq)
	elem = stm.FetchElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	ping := xmpp.NewElementNamespace("ping", pingNamespace)
	iq.AppendElement(ping)

	x.ProcessIQ(iq)
	elem = stm.FetchElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	iq.SetType(xmpp.GetType)
	x.ProcessIQ(iq)
	elem = stm.FetchElement()
	require.Equal(t, iqID, elem.ID())
}

func TestXEP0199_SendPing(t *testing.T) {
	t.Parallel()
	j1, _ := jid.New("ortuman", "jackal.im", "balcony", true)

	stm := stream.NewMockC2S("abcd", j1)
	defer stm.Disconnect(nil)

	stm.SetUsername("ortuman")

	x := New(&Config{Send: true, SendInterval: 1}, stm)

	x.StartPinging()

	// wait for ping...
	elem := stm.FetchElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.NotNil(t, elem.Elements().ChildNamespace("ping", pingNamespace))

	// send pong...
	x.ProcessIQ(xmpp.NewIQType(elem.ID(), xmpp.ResultType))
	x.ResetDeadline()

	// wait next ping...
	elem = stm.FetchElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.NotNil(t, elem.Elements().ChildNamespace("ping", pingNamespace))

	// expect disconnection...
	err := stm.WaitDisconnection()
	require.NotNil(t, err)
}

func TestXEP0199_Disconnect(t *testing.T) {
	t.Parallel()
	j1, _ := jid.New("ortuman", "jackal.im", "balcony", true)

	stm := stream.NewMockC2S("abcd", j1)
	defer stm.Disconnect(nil)

	stm.SetUsername("ortuman")

	x := New(&Config{Send: true, SendInterval: 1}, stm)

	x.StartPinging()

	// wait next ping...
	elem := stm.FetchElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.NotNil(t, elem.Elements().ChildNamespace("ping", pingNamespace))

	// expect disconnection...
	err := stm.WaitDisconnection()
	require.NotNil(t, err)
	require.Equal(t, "connection-timeout", err.Error())
}
