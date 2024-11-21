/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupRead,
		Schema: map[string]*schema.Schema{
			"attributes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Attributes for this group",
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
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Metadata of group object creation",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Display name of this group",
			},
			"id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The unique group ID",
			},
			"users": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Members of this group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Display name of this user",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique user ID",
						},
					},
				},
			},
		},
	}
}

func dataSourceGroupUsersParser(llusers []LldapUser) []map[string]any {
	result := make([]map[string]any, len(llusers))
	for i, lluser := range llusers {
		group := map[string]any{
			"id":           lluser.Id,
			"display_name": lluser.DisplayName,
		}
		result[i] = group
	}
	return result
}

func dataSourceGroupRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Get("id").(int)
	lc := m.(*LldapClient)
	llgroup, getGroupErr := lc.GetGroup(id)
	if getGroupErr != nil {
		return getGroupErr
	}
	d.SetId(strconv.Itoa(llgroup.Id))
	for k, v := range map[string]interface{}{
		"display_name":  llgroup.DisplayName,
		"creation_date": llgroup.CreationDate,
		"users":         dataSourceGroupUsersParser(llgroup.Users),
		"attributes":    attributesParser(llgroup.Attributes),
	} {
		if setErr := d.Set(k, v); setErr != nil {
			return diag.FromErr(setErr)
		}
	}
	return nil
}
