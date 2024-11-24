/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Description: "Reads a LLDAP user, with group memberships",
		Schema: map[string]*schema.Schema{
			"attributes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Custom attributes for this user",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unique name of this attribute",
						},
						"value": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of values for this attribute",
						},
					},
				},
			},
			"avatar": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Base 64 encoded JPEG image",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Metadata of user object creation",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Display name of this user",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique user email",
			},
			"first_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "First name of this user",
			},
			"groups": {
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
			},
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique user ID",
				StateFunc: func(val any) string {
					return strings.ToLower(val.(string))
				},
			},
			"last_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last name of this user",
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique username",
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID of user",
			},
		},
	}
}

func dataSourceUserRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Get("id").(string)
	lc := m.(*LldapClient)
	user, getUserErr := lc.GetUser(id)
	if getUserErr != nil {
		return getUserErr
	}
	d.SetId(user.Id)
	for k, v := range map[string]interface{}{
		"attributes":    attributesParser(user.Attributes),
		"avatar":        user.Avatar,
		"creation_date": user.CreationDate,
		"display_name":  user.DisplayName,
		"email":         user.Email,
		"first_name":    user.FirstName,
		"groups":        dataSourceGroupsParser(user.Groups),
		"last_name":     user.LastName,
		"username":      user.Id,
		"uuid":          user.Uuid,
	} {
		if setErr := d.Set(k, v); setErr != nil {
			return diag.FromErr(setErr)
		}
	}
	return nil
}
