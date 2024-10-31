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
