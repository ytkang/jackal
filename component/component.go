/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package component

import (
	"sync"

	"fmt"

	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/xmpp"
)

type Component interface {
	Domain() string
	ProcessStanza(stanza xmpp.Stanza)
	Shutdown()
}

type Config struct {
}

type component struct {
	mu    sync.RWMutex
	comps map[string]Component
}

// singleton interface
var (
	instMu      sync.RWMutex
	inst        *component
	initialized bool
)

// Initialize initializes the components manager.
func Initialize(cfg *Config) {
	instMu.Lock()
	defer instMu.Unlock()
	if initialized {
		return
	}
	inst = &component{
		comps: make(map[string]Component),
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
	for _, comp := range inst.comps {
		comp.Shutdown()
	}
	inst = nil
	initialized = false
}

func IsComponentDomain(domain string) bool {
	return instance().isComponentDomain(domain)
}

func RegisteredDomains() []string {
	return instance().registeredDomains()
}

func Register(comp Component) error {
	return instance().register(comp)
}

func Unregister(comp Component) error {
	return instance().unregister(comp)
}

func instance() *component {
	instMu.RLock()
	defer instMu.RUnlock()
	if inst == nil {
		log.Fatalf("component manager not initialized")
	}
	return inst
}

func (c *component) registeredDomains() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var ret []string
	for _, comp := range c.comps {
		ret = append(ret, comp.Domain())
	}
	return ret
}

func (c *component) isComponentDomain(domain string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.comps[domain]
	return ok
}

func (c *component) register(comp Component) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.comps[comp.Domain()]; ok {
		return fmt.Errorf("component: domain %s already registered", comp.Domain())
	}
	c.comps[comp.Domain()] = comp
	return nil
}

func (c *component) unregister(comp Component) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.comps[comp.Domain()]; !ok {
		return fmt.Errorf("component: domain %s not registered", comp.Domain())
	}
	delete(c.comps, comp.Domain())
	return nil
}
