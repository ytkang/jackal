/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package httpupload

import (
	"errors"
	"time"
)

const (
	defaultUploadPath = "/var/lib/jackal/httpupload"
	defaultSizeLimit  = 1048576
)

type Config struct {
	Host        string
	BaseURL     string
	Port        int
	UploadPath  string
	SizeLimit   int
	Quota       int
	ExpireAfter time.Duration
}

type configProxy struct {
	Host        string `yaml:"host"`
	BaseURL     string `yaml:"base_url"`
	Port        int    `yaml:"port"`
	UploadPath  string `yaml:"upload_path"`
	SizeLimit   int    `yaml:"size_limit"`
	Quota       int    `yaml:"quota"`
	ExpireAfter int    `yaml:"expire_after"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	p := configProxy{}
	if err := unmarshal(&p); err != nil {
		return err
	}
	// mandatory fields
	if len(p.Host) == 0 {
		return errors.New("httpupload.Config: host value must be set")
	}
	if len(p.BaseURL) == 0 {
		return errors.New("httpupload.Config: base_url must be set")
	}
	if p.Port == 0 {
		return errors.New("httpupload.Config: port value must be set")
	}
	c.Host = p.Host
	c.BaseURL = p.BaseURL
	c.Port = p.Port

	// optional fields
	c.UploadPath = p.UploadPath
	if len(c.UploadPath) == 0 {
		c.UploadPath = defaultUploadPath
	}
	c.SizeLimit = p.SizeLimit
	if c.SizeLimit == 0 {
		c.SizeLimit = defaultSizeLimit
	}
	c.Quota = p.Quota
	c.ExpireAfter = time.Second * time.Duration(p.ExpireAfter)
	return nil
}
