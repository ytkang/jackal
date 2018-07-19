/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package http_upload

type Config struct {
	Domain string
}

type HttpUpload struct {
	cfg *Config
}

func New(cfg *Config) *HttpUpload {
	h := &HttpUpload{cfg: cfg}
	return h
}
