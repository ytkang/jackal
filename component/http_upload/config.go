/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package http_upload

import "time"

type Config struct {
	Host        string
	BaseURL     string
	Port        int
	UploadPath  string
	SizeLimit   int
	Quota       int
	ExpireAfter time.Duration
}
