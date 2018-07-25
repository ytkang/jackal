/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package httpupload

import "time"

type Config struct {
	Host        string        `yaml:"host"`
	BaseURL     string        `yaml:"base_url"`
	Port        int           `yaml:"port"`
	UploadPath  string        `yaml:"upload_path"`
	SizeLimit   int           `yaml:"size_limit"`
	Quota       int           `yaml:"quota"`
	ExpireAfter time.Duration `yaml:"expire_after"`
}
