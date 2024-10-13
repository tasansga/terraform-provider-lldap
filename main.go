package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	lldap "github.com/tasansga/tf-provider-lldap/lldap"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: lldap.Provider,
	})
}
