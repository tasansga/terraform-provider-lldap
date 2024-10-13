package lldap

import (
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

type Config struct {
	Url      *url.URL
	UserName string
	Password string
}

func (c *Config) Client() (*LldapClient, diag.Diagnostics) {
	return &LldapClient{
		Token:  "",
		Config: c,
	}, nil
}
