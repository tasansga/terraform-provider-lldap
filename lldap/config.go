/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"net/url"
)

type Config struct {
	Context               context.Context
	HttpUrl               *url.URL
	LdapUrl               *url.URL
	UserName              string
	Password              string
	InsecureSkipCertCheck bool
	BaseDn                string
}
