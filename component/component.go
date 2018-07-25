/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package component

import (
	"sync"

	"github.com/ortuman/jackal/component/httpupload"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/xmpp"
)

type Component interface {
	Host() string
	ProcessStanza(stanza xmpp.Stanza)
	Shutdown()
}

// singleton interface
var (
	instMu      sync.RWMutex
	comps       map[string]Component
	initialized bool
)

// Initialize initializes the components manager.
func Initialize(cfg *Config) {
	instMu.Lock()
	defer instMu.Unlock()
	if initialized {
		return
	}
	comps = make(map[string]Component)

	if cfg.HttpUpload != nil {
		comp := httpupload.New(cfg.HttpUpload)
		comps[comp.Host()] = comp
		log.Infof("'http_upload' component enabled [host: %s, port: %d]", comp.Host(), cfg.HttpUpload.Port)
	}
	initialized = true
}

// Shutdown shuts down components manager system.
// This method should be used only for testing purposes.
func Shutdown() {
	instMu.Lock()
	defer instMu.Unlock()
	if !initialized {
		return
	}
	for _, comp := range comps {
		comp.Shutdown()
	}
	comps = nil
	initialized = false
}
