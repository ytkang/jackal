/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package component

import "github.com/ortuman/jackal/component/httpupload"

type Config struct {
	HttpUpload *httpupload.Config `yaml:"http_upload"`
}
