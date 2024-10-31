package lldap

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_dn": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "dc=example,dc=com",
				Description: "Base DN",
			},
			"http_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LLDAP_HTTP_URL", nil),
				Description: "HTTP URL in the format http[s]://(hostname)[:port]",
			},
			"insecure_skip_cert_check": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable check for valid certificate chain for https/ldaps",
			},
			"ldap_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LLDAP_LDAP_URL", nil),
				Description: "LDAP URL in the format ldap[s]://(hostname)[:port]",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LLDAP_PASSWORD", nil),
				Description: "admin account password",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LLDAP_USERNAME", nil),
				Description: "admin account username",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"lldap_group":  resourceGroup(),
			"lldap_member": resourceMember(),
			"lldap_user":   resourceUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"lldap_group":  dataSourceGroup(),
			"lldap_groups": dataSourceGroups(),
			"lldap_user":   dataSourceUser(),
			"lldap_users":  dataSourceUsers(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		rawHttpUrl := d.Get("http_url").(string)
		parsedHttpUrl, parseHttpUrlErr := url.Parse(rawHttpUrl)
		if parseHttpUrlErr != nil {
			return nil, diag.FromErr(parseHttpUrlErr)
		}
		if parsedHttpUrl.Scheme != "http" && parsedHttpUrl.Scheme != "https" {
			return nil, diag.Errorf("Invalid LLDAP HTTP URL: '%s'", rawHttpUrl)
		}
		rawLdapUrl := d.Get("ldap_url").(string)
		parsedLdapUrl, parseLdapUrlErr := url.Parse(rawLdapUrl)
		if parseLdapUrlErr != nil {
			return nil, diag.FromErr(parseLdapUrlErr)
		}
		if parsedLdapUrl.Scheme != "ldap" && parsedLdapUrl.Scheme != "ldaps" {
			return nil, diag.Errorf("Invalid LLDAP LDAP URL: '%s'", rawLdapUrl)
		}
		client := LldapClient{
			Config: Config{
				Context:  ctx,
				HttpUrl:  parsedHttpUrl,
				LdapUrl:  parsedLdapUrl,
				UserName: d.Get("username").(string),
				Password: d.Get("password").(string),
				BaseDn:   d.Get("base_dn").(string),
			},
		}
		return &client, nil
	}

	return provider
}
