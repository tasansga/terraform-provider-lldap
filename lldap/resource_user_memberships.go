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

func resourceUserMemberships() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserMembershipsCreate,
		ReadContext:   resourceUserMembershipsRead,
		UpdateContext: resourceUserMembershipsUpdate,
		DeleteContext: resourceUserMembershipsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
				_ = d.Set("id", d.Id())
				return schema.ImportStatePassthroughContext(ctx, d, m)
			},
		},
		Description: "Exclusively manages all LLDAP memberhips for this specific user",
		Schema: map[string]*schema.Schema{
			"group_ids": {
				Type:        schema.TypeSet,
				Description: "Groups id where the user must be a member",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: func(val any, key string) (warns []string, errs []error) {
						v := val.(string)
						if _, err := strconv.Atoi(v); err != nil {
							errs = append(errs, fmt.Errorf("%q must be a string representing an integer, got: %s", key, v))
						}
						return
					},
				},
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID representing this specific user memberships",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique user id",
			},
		},
	}
}

func resourceUserMembershipsSetResourceData(d *schema.ResourceData, user *LldapUser) diag.Diagnostics {
	groupIdsStr := make([]string, len(user.Groups))
	for i, group := range user.Groups {
		groupIdsStr[i] = strconv.Itoa(group.Id)
	}
	slices.Sort(groupIdsStr)
	for k, v := range map[string]any{
		"user_id":   user.Id,
		"group_ids": groupIdsStr,
	} {
		if v != nil {
			if setErr := d.Set(k, v); setErr != nil {
				return diag.FromErr(setErr)
			}
		}
	}
	return nil
}

func resourceUserMembershipsGetGroupIds(d *schema.ResourceData) ([]int, diag.Diagnostics) {
	groupIdsUnparsedList := d.Get("group_ids").(*schema.Set).List()
	groupIds := make([]int, len(groupIdsUnparsedList))
	for i, groupIdVal := range groupIdsUnparsedList {
		groupId, groupIdParseErr := strconv.Atoi(groupIdVal.(string))
		if groupIdParseErr != nil {
			return nil, diag.FromErr(groupIdParseErr)
		}
		groupIds[i] = groupId
	}
	return groupIds, nil
}

func resourceUserMembershipsCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	userId := d.Get("user_id").(string)
	groupIds, getGroupIdsErr := resourceUserMembershipsGetGroupIds(d)
	if getGroupIdsErr != nil {
		return getGroupIdsErr
	}
	lc := m.(*LldapClient)
	user, getUserErr := lc.GetUser(userId)
	if getUserErr != nil {
		return getUserErr
	}
	for _, groupId := range groupIds {
		addErr := lc.AddUserToGroup(groupId, userId)
		if addErr != nil {
			return addErr
		}
	}
	d.SetId(user.Id)
	return nil
}

func resourceUserMembershipsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	userId := d.Get("user_id").(string)
	user, getUserErr := lc.GetUser(userId)
	if getUserErr != nil {
		// If the user was not found, mark the resource as deleted so Terraform will recreate it
		if isEntityNotFoundError(getUserErr) {
			d.SetId("")
			return nil
		}
		return getUserErr
	}
	setRdErr := resourceUserMembershipsSetResourceData(d, user)
	if setRdErr != nil {
		return setRdErr
	}
	return nil
}

func resourceUserMembershipsUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	userId := d.Get("user_id").(string)
	userWantsGroupIds, getGroupIdsErr := resourceUserMembershipsGetGroupIds(d)
	if getGroupIdsErr != nil {
		return getGroupIdsErr
	}
	lc := m.(*LldapClient)
	user, getUserErr := lc.GetUser(userId)
	if getUserErr != nil {
		return getUserErr
	}
	userHasGroupIds := user.GetGroupIds()
	for _, wantsGroupId := range userWantsGroupIds {
		if !slices.Contains(userHasGroupIds, wantsGroupId) {
			addErr := lc.AddUserToGroup(wantsGroupId, user.Id)
			if addErr != nil {
				return addErr
			}
		}
	}
	for _, hasGroupId := range userHasGroupIds {
		if !slices.Contains(userWantsGroupIds, hasGroupId) {
			removeErr := lc.RemoveUserFromGroup(hasGroupId, user.Id)
			if removeErr != nil {
				return removeErr
			}
		}
	}
	return nil
}

func resourceUserMembershipsDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	userId := d.Get("user_id").(string)
	groupIds, getGroupIdsErr := resourceUserMembershipsGetGroupIds(d)
	if getGroupIdsErr != nil {
		return getGroupIdsErr
	}
	lc := m.(*LldapClient)
	for _, groupId := range groupIds {
		removeErr := lc.RemoveUserFromGroup(groupId, userId)
		if removeErr != nil {
			return removeErr
		}
	}
	return nil
}
