/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceMemberIdSeparator = ":"

func resourceMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMemberCreate,
		ReadContext:   resourceMemberRead,
		DeleteContext: resourceMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				id := d.Id()
				_, _, found := strings.Cut(id, ResourceMemberIdSeparator)
				if !found {
					return nil, fmt.Errorf("not a valid member id: %s", id)
				}
				_ = d.Set("id", d.Id())
				return schema.ImportStatePassthroughContext(ctx, d, m)
			},
		},
		Description: "Manages a LLDAP memberhip, i.e. a group-user relationship",
		Schema: map[string]*schema.Schema{
			"group_display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Display name of this group",
			},
			"group_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The unique group ID",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The member 'ID', constructed as group_id:user_id",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique user ID",
			},
		},
	}
}

func resourceMemberGetId(groupId int, userId string) string {
	id := fmt.Sprintf("%d%s%s", groupId, ResourceMemberIdSeparator, userId)
	return id
}

func resourceMemberCreate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	groupId := d.Get("group_id").(int)
	userId := d.Get("user_id").(string)
	id := resourceMemberGetId(groupId, userId)
	lc := m.(*LldapClient)
	createErr := lc.AddUserToGroup(groupId, userId)
	if createErr != nil {
		return createErr
	}
	group, getGroupErr := lc.GetGroup(groupId)
	if getGroupErr != nil {
		return getGroupErr
	}
	d.SetId(id)
	if setErr := d.Set("group_display_name", group.DisplayName); setErr != nil {
		return diag.FromErr(setErr)
	}
	return nil
}

func resourceMemberRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Id()
	groupIdString, userId, ok := strings.Cut(id, ResourceMemberIdSeparator)
	if !ok {
		return diag.Errorf("not a valid lldap_member id: %s", id)
	}

	groupId, err := strconv.Atoi(groupIdString)
	if err != nil {
		return diag.Errorf("group_id should be an integer: %v", err)
	}

	lc := m.(*LldapClient)
	group, getGroupErr := lc.GetGroup(groupId)
	if getGroupErr != nil {
		return getGroupErr
	}
	groupMembers := make([]string, len(group.Users))
	for _, user := range group.Users {
		groupMembers = append(groupMembers, user.Id)
	}
	if !slices.Contains(groupMembers, userId) {
		return diag.Errorf("User not a member of group!")
	}
	return nil
}

func resourceMemberDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	groupId := d.Get("group_id").(int)
	userId := d.Get("user_id").(string)
	lc := m.(*LldapClient)
	removeErr := lc.RemoveUserFromGroup(groupId, userId)
	if removeErr != nil {
		return removeErr
	}
	return nil
}
