package lldap

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"lldap_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LLDAP_URL", nil),
				Description: "LLDAP URL in the format http[s]://(hostname)[:port]",
			},
			"lldap_username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LLDAP_USERNAME", nil),
				Description: "LLDAP admin account username",
			},
			"lldap_password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LLDAP_PASSWORD", nil),
				Description: "LLDAP admin account password",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"lldap_group": resourceGroup(),
			"lldap_user":  resourceUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"lldap_group":  dataSourceGroup(),
			"lldap_groups": dataSourceGroups(),
			"lldap_user":   dataSourceUser(),
			"lldap_users":  dataSourceUsers(),
		},
	}

	provider.ConfigureContextFunc = func(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		config := Config{
			Url:      d.Get("lldap_url").(string),
			UserName: d.Get("lldap_username").(string),
			Password: d.Get("lldap_password").(string),
		}
		return config.Client()
	}
	return provider
}
