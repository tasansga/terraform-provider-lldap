package lldap

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

type Config struct {
	Url      string
	UserName string
	Password string
}

func (c *Config) Client() (*LldapClient, diag.Diagnostics) {
	return &LldapClient{
		Token:  "",
		Config: c,
	}, nil
}
