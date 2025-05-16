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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGroupMemberships() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupMembershipsCreate,
		ReadContext:   resourceGroupMembershipsRead,
		UpdateContext: resourceGroupMembershipsUpdate,
		DeleteContext: resourceGroupMembershipsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
				_ = d.Set("id", d.Id())
				return schema.ImportStatePassthroughContext(ctx, d, m)
			},
		},
		Description: "Exclusively manages all LLDAP memberhips for this specific group",
		Schema: map[string]*schema.Schema{
			"user_ids": {
				Type:        schema.TypeSet,
				Description: "User ids that must be members of this group",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID representing this specific group memberships",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique group id",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if _, err := strconv.Atoi(v); err != nil {
						errs = append(errs, fmt.Errorf("%q must be a string representing an integer, got: %s", key, v))
					}
					return
				},
			},
		},
	}
}

func resourceGroupMembershipsSetResourceData(d *schema.ResourceData, group *LldapGroup) diag.Diagnostics {
	userIds := make([]string, len(group.Users))
	for i, user := range group.Users {
		userIds[i] = user.Id
	}
	slices.Sort(userIds)
	for k, v := range map[string]any{
		"group_id": strconv.Itoa(group.Id),
		"user_ids": userIds,
	} {
		if v != nil {
			if setErr := d.Set(k, v); setErr != nil {
				return diag.FromErr(setErr)
			}
		}
	}
	return nil
}

func resourceGroupMembershipsGetUserIds(d *schema.ResourceData) ([]string, diag.Diagnostics) {
	userIdsUnparsedList := d.Get("user_ids").(*schema.Set).List()
	userIds := make([]string, len(userIdsUnparsedList))
	for i, userIdVal := range userIdsUnparsedList {
		userIds[i] = userIdVal.(string)
	}
	return userIds, nil
}

func resourceGroupMembershipsCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	groupId := d.Get("group_id").(string)
	userIds, getUserIdsErr := resourceGroupMembershipsGetUserIds(d)
	if getUserIdsErr != nil {
		return getUserIdsErr
	}
	lc := m.(*LldapClient)
	groupIdInt, _ := strconv.Atoi(groupId)
	for _, userId := range userIds {
		addErr := lc.AddUserToGroup(groupIdInt, userId)
		if addErr != nil {
			return addErr
		}
	}
	d.SetId(groupId)
	return nil
}

func resourceGroupMembershipsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	groupId := d.Get("group_id").(string)
	groupIdInt, _ := strconv.Atoi(groupId)
	lc := m.(*LldapClient)
	group, getGroupErr := lc.GetGroup(groupIdInt)
	if getGroupErr != nil {
		return getGroupErr
	}
	setRdErr := resourceGroupMembershipsSetResourceData(d, group)
	if setRdErr != nil {
		return setRdErr
	}
	return nil
}

func resourceGroupMembershipsUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	groupId := d.Get("group_id").(string)
	groupIdInt, _ := strconv.Atoi(groupId)
	groupWantsUserIds, getGroupIdsErr := resourceGroupMembershipsGetUserIds(d)
	if getGroupIdsErr != nil {
		return getGroupIdsErr
	}
	lc := m.(*LldapClient)
	group, getGroupErr := lc.GetGroup(groupIdInt)
	if getGroupErr != nil {
		return getGroupErr
	}
	groupHasUserIds := group.GetUserIds()
	for _, wantsUserId := range groupWantsUserIds {
		if !slices.Contains(groupHasUserIds, wantsUserId) {
			addErr := lc.AddUserToGroup(group.Id, wantsUserId)
			if addErr != nil {
				return addErr
			}
		}
	}
	for _, hasUserId := range groupHasUserIds {
		if !slices.Contains(groupWantsUserIds, hasUserId) {
			removeErr := lc.RemoveUserFromGroup(group.Id, hasUserId)
			if removeErr != nil {
				return removeErr
			}
		}
	}
	return nil
}

func resourceGroupMembershipsDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	groupId := d.Get("group_id").(string)
	groupIdInt, _ := strconv.Atoi(groupId)
	userIds, getUserIdsErr := resourceGroupMembershipsGetUserIds(d)
	if getUserIdsErr != nil {
		return getUserIdsErr
	}
	lc := m.(*LldapClient)
	for _, userId := range userIds {
		removeErr := lc.RemoveUserFromGroup(groupIdInt, userId)
		if removeErr != nil {
			return removeErr
		}
	}
	return nil
}
