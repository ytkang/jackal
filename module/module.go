/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package module

import (
	"sync"

	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/module/xep0092"
	"github.com/ortuman/jackal/xmpp"
)

// Module represents a generic XMPP module.
type Module interface {
	// Features returns disco entity features associated to the module.
	Features() []string
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

type Mods struct {
	DiscoInfo *xep0030.DiscoInfo
	Version   *xep0092.Version

	iqHandlers []IQHandler
}

var (
	instMu      sync.RWMutex
	mods        Mods
	shutdownCh  chan struct{}
	initialized bool
)

func Initialize(cfg *Config) {
	instMu.Lock()
	defer instMu.Unlock()
	if initialized {
		return
	}
	initializeModules(cfg)
	initialized = true
}

func Shutdown() {
	instMu.Lock()
	defer instMu.Unlock()
	if !initialized {
		return
	}
	close(shutdownCh)
	mods = Mods{}
	initialized = false
}

func Modules() Mods {
	return mods
}

func IQHandlers() []IQHandler {
	return mods.iqHandlers
}

func initializeModules(cfg *Config) {
	shutdownCh = make(chan struct{})
	mods.DiscoInfo = xep0030.New(shutdownCh)

	// XEP-0092: Software Version (https://xmpp.org/extensions/xep-0092.html)
	if _, ok := cfg.Enabled["version"]; ok {
		mods.Version = xep0092.New(&cfg.Version, shutdownCh)
		mods.iqHandlers = append(mods.iqHandlers, mods.Version)
	}
}
