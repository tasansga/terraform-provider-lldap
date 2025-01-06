/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const resourceUserAttributeAssignmentIdSeparator = ":"

func resourceUserAttributeAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserAttributeAssignmentCreate,
		ReadContext:   resourceUserAttributeAssignmentRead,
		DeleteContext: resourceUserAttributeAssignmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				id := d.Id()
				_, _, found := strings.Cut(id, resourceUserAttributeAssignmentIdSeparator)
				if !found {
					return nil, fmt.Errorf("not a valid attribute assignment id: %s", id)
				}
				_ = d.Set("id", d.Id())
				return schema.ImportStatePassthroughContext(ctx, d, m)
			},
		},
		Description: "Manage a custom attribute assignment for an user",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The assignment 'ID', constructed as user_id:attribute_name",
			},
			"attribute_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The attribute name",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique user ID",
			},
			"value": {
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true,
				Description: "The value(s) for this attribute",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceUserAttributeAssignmentCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	userId := d.Get("user_id").(string)
	attributeId := d.Get("attribute_id").(string)
	valueRaw := d.Get("value").(*schema.Set)
	valueRawList := valueRaw.List()
	value := make([]string, len(valueRawList))
	for i, vRaw := range valueRawList {
		value[i] = vRaw.(string)
	}
	something, _ := json.Marshal(value)
	tflog.Error(ctx, fmt.Sprintf("Got something: %s", string(something)))
	id := fmt.Sprintf("%s%s%s", userId, resourceUserAttributeAssignmentIdSeparator, attributeId)
	tflog.Debug(ctx, fmt.Sprintf("Will create user attribute assignment with id: %s", id))
	d.SetId(id)
	lc := m.(*LldapClient)
	addAttrErr := lc.AddAttributeToUser(userId, attributeId, []string{})
	if addAttrErr != nil {
		return addAttrErr
	}
	tflog.Info(ctx, fmt.Sprintf("Created user attribute assignment with id: %s", id))
	return nil
}

func resourceUserAttributeAssignmentRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Id()
	userId, attributeId, ok := strings.Cut(id, resourceUserAttributeAssignmentIdSeparator)
	if !ok {
		return diag.Errorf("not a valid lldap_user_attribute_assignment id: %s", id)
	}

	lc := m.(*LldapClient)
	user, getUserErr := lc.GetUser(userId)
	if getUserErr != nil {
		return getUserErr
	}
	userAttributes := make([]string, len(user.Attributes))
	var value []string
	for _, attr := range user.Attributes {
		userAttributes = append(userAttributes, attr.Name)
		if attr.Name == attributeId {
			value = attr.Value
		}
	}
	if !slices.Contains(userAttributes, attributeId) {
		return diag.Errorf("User is missing attribute!")
	}
	for k, v := range map[string]interface{}{
		"user_id":      userId,
		"attribute_id": attributeId,
		"value":        value,
	} {
		if setErr := d.Set(k, v); setErr != nil {
			return diag.FromErr(setErr)
		}
	}
	return nil
}

func resourceUserAttributeAssignmentDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	userId := d.Get("user_id").(string)
	attributeId := d.Get("attribute_id").(string)
	lc := m.(*LldapClient)
	removeAttrErr := lc.RemoveAttributeFromUser(userId, attributeId)
	if removeAttrErr != nil {
		return removeAttrErr
	}
	return nil
}
