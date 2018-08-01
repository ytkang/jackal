/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package module

import (
	"sync"

	"github.com/ortuman/jackal/xmpp"
)

// Module represents a generic XMPP module.
type Module interface {
	// Features returns disco entity features associated to the module.
	Features() []string

	// Shutdown closes module.
	Shutdown()
}

// IQHandler represents an IQ handler module.
type IQHandler interface {
	Module

	// MatchesIQ returns whether or not an IQ should be
	// processed by the module.
	MatchesIQ(iq *xmpp.IQ) bool

	// ProcessIQ processes a module IQ taking according actions
	// over the associated stream.
	ProcessIQ(iq *xmpp.IQ)
}

type Modules struct {
}

var (
	my          sync.RWMutex
	mods        Modules
	iqHandlers  []IQHandler
	initialized bool
)

func Initialize(cfg *Config) {
}

func Shutdown() {
}

func All() Modules {
	return mods
}

func IQHandlers() []IQHandler {
	return iqHandlers
}
