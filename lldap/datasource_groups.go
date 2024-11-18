/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupsRead,
		Schema: map[string]*schema.Schema{
			"groups": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of all groups",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"creation_date": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Metadata of group object creation",
						},
						"display_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Display name of this group",
						},
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique group ID",
						},
					},
				},
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Generated ID representing the groups",
			},
		},
	}
}

func dataSourceGroupsRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	llgroups, getGroupsErr := lc.GetGroups()
	if getGroupsErr != nil {
		return getGroupsErr
	}
	dataSourceSetHashId(d, llgroups)
	if setErr := d.Set("groups", dataSourceGroupsParser(llgroups)); setErr != nil {
		return diag.Errorf("could not create group set: %s", setErr)
	}
	return nil
}
