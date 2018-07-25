/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package http_upload

import (
	"github.com/ortuman/jackal/xmpp"
)

type HttpUpload struct {
	cfg *Config
}

func New(cfg *Config) *HttpUpload {
	h := &HttpUpload{cfg: cfg}
	return h
}

func (c *HttpUpload) Host() string {
	return c.cfg.Host
}

func (c *HttpUpload) ProcessStanza(stanza xmpp.Stanza) {
}

func (c *HttpUpload) Shutdown() {
}
