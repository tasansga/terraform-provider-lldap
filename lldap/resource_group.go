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

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				_ = d.Set("id", d.Id())
				return schema.ImportStatePassthroughContext(ctx, d, m)
			},
		},
		Schema: map[string]*schema.Schema{
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Metadata of group object creation",
			},
			"attributes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Attributes for this group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name of this attribute",
						},
						"value": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of values for this attribute",
						},
					},
				},
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Display name of this group",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique group ID",
			},
			"users": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of users who are members of this group",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID of group",
			},
		},
	}
}

func resourceGroupSetResourceData(d *schema.ResourceData, group *LldapGroup) diag.Diagnostics {
	for k, v := range map[string]interface{}{
		"attributes":    dataSourceAttributesParser(group.Attributes),
		"creation_date": group.CreationDate,
		"display_name":  group.DisplayName,
		"users":         resourceGroupUsersParser(group.Users),
		"uuid":          group.Uuid,
	} {
		if setErr := d.Set(k, v); setErr != nil {
			return diag.FromErr(setErr)
		}
	}
	return nil
}

func resourceGroupUsersParser(llusers []LldapUser) *schema.Set {
	result := make([]interface{}, len(llusers))
	for i, lluser := range llusers {
		result[i] = lluser.Id
	}
	return schema.NewSet(schema.HashString, result)
}

func resourceGroupCreate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	group := LldapGroup{
		DisplayName: d.Get("display_name").(string),
	}
	lc := m.(*LldapClient)
	createErr := lc.CreateGroup(&group)
	if createErr != nil {
		return createErr
	}
	d.SetId(strconv.Itoa(group.Id))
	setRdErr := resourceGroupSetResourceData(d, &group)
	if setRdErr != nil {
		return setRdErr
	}
	return nil
}

func resourceGroupRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	groupId, getGroupIdErr := strconv.Atoi(d.Id())
	if getGroupIdErr != nil {
		return diag.FromErr(getGroupIdErr)
	}
	group, getGroupErr := lc.GetGroup(groupId)
	if getGroupErr != nil {
		return getGroupErr
	}
	setRdErr := resourceGroupSetResourceData(d, group)
	if setRdErr != nil {
		return setRdErr
	}
	return nil
}

func resourceGroupUpdate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	groupId, getGroupIdErr := strconv.Atoi(d.Id())
	if getGroupIdErr != nil {
		return diag.FromErr(getGroupIdErr)
	}
	displayName := d.Get("display_name").(string)
	updateErr := lc.UpdateGroupDisplayName(groupId, displayName)
	if updateErr != nil {
		return updateErr
	}
	return nil
}

func resourceGroupDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	groupId, getGroupIdErr := strconv.Atoi(d.Id())
	if getGroupIdErr != nil {
		return diag.FromErr(getGroupIdErr)
	}
	deleteErr := lc.DeleteGroup(groupId)
	if deleteErr != nil {
		return deleteErr
	}
	return nil
}
