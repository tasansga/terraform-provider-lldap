/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strconv"

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
				Description: "Base DN, defaults to `dc=example,dc=com`",
			},
			"http_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LLDAP_HTTP_URL", nil),
				Description: "HTTP URL in the format `http[s]://(hostname)[:port]`, can be set using the `LLDAP_HTTP_URL` environment variable",
			},
			"insecure_skip_cert_check": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable check for valid certificate chain for https/ldaps (default: `false`)",
			},
			"ldap_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LLDAP_LDAP_URL", nil),
				Description: "LDAP URL in the format `ldap[s]://(hostname)[:port]`, can be set using the `LLDAP_LDAP_URL` environment variable",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LLDAP_PASSWORD", nil),
				Description: "admin account password, can be set using the `LLDAP_PASSWORD` environment variable",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "admin",
				Description: "admin account username, defaults to `admin`",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"lldap_group_attribute_assignment": resourceGroupAttributeAssignment(),
			"lldap_group_attribute":            resourceGroupAttribute(),
			"lldap_group_memberships":          resourceGroupMemberships(),
			"lldap_group":                      resourceGroup(),
			"lldap_member":                     resourceMember(),
			"lldap_user_attribute_assignment":  resourceUserAttributeAssignment(),
			"lldap_user_attribute":             resourceUserAttribute(),
			"lldap_user_memberships":           resourceUserMemberships(),
			"lldap_user":                       resourceUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"lldap_group_attributes": dataSourceGroupAttributes(),
			"lldap_group":            dataSourceGroup(),
			"lldap_groups":           dataSourceGroups(),
			"lldap_user_attributes":  dataSourceUserAttributes(),
			"lldap_user":             dataSourceUser(),
			"lldap_users":            dataSourceUsers(),
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

func dataSourceSetHashId(d *schema.ResourceData, v any) diag.Diagnostics {
	hashBase, marshalErr := json.Marshal(v)
	if marshalErr != nil {
		return diag.FromErr(marshalErr)
	}
	hash := sha1.New()
	hash.Write([]byte(hashBase))
	hashString := hex.EncodeToString(hash.Sum(nil))
	d.SetId(hashString)
	return nil
}

func dataSourceGroupsParser(llgroups []LldapGroup) []map[string]any {
	result := make([]map[string]any, len(llgroups))
	for i, llgroup := range llgroups {
		group := map[string]any{
			"id":            strconv.Itoa(llgroup.Id),
			"display_name":  llgroup.DisplayName,
			"creation_date": llgroup.CreationDate,
		}
		result[i] = group
	}
	return result
}

func attributesParser(attrs []LldapCustomAttribute) []map[string]any {
	result := make([]map[string]any, len(attrs))
	for i, llattr := range attrs {
		attr := map[string]any{
			"name":  llattr.Name,
			"value": llattr.Value,
		}
		result[i] = attr
	}
	return result
}

var dataSourceGroupsSchema = schema.Schema{
	Type:        schema.TypeSet,
	Computed:    true,
	Description: "Groups where the user is a member",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"creation_date": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Metadata of group object creation",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Display name of the group",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique group ID",
			},
		},
	},
}
