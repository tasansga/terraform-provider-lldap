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
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const resourceGroupAttributeAssignmentIdSeparator = ":"

func resourceGroupAttributeAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupAttributeAssignmentCreate,
		ReadContext:   resourceGroupAttributeAssignmentRead,
		UpdateContext: resourceGroupAttributeAssignmentUpdate,
		DeleteContext: resourceGroupAttributeAssignmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
				id := d.Id()
				_, _, found := strings.Cut(id, resourceGroupAttributeAssignmentIdSeparator)
				if !found {
					return nil, fmt.Errorf("not a valid attribute assignment id: %s", id)
				}
				_ = d.Set("id", d.Id())
				return schema.ImportStatePassthroughContext(ctx, d, m)
			},
		},
		Description: "Manage a custom attribute assignment for a group",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The assignment 'ID', constructed as group_id:attribute_name",
			},
			"attribute_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The attribute name",
			},
			"group_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The unique group ID",
			},
			"value": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The value(s) for this attribute",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceGroupAttributeAssignmentCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	groupId := d.Get("group_id").(int)
	attributeId := d.Get("attribute_id").(string)
	valueRaw := d.Get("value").(*schema.Set)
	valueRawList := valueRaw.List()
	value := make([]string, len(valueRawList))
	for i, vRaw := range valueRawList {
		value[i] = vRaw.(string)
	}
	something, _ := json.Marshal(value)
	tflog.Error(ctx, fmt.Sprintf("Got something: %s", string(something)))
	id := fmt.Sprintf("%d%s%s", groupId, resourceGroupAttributeAssignmentIdSeparator, attributeId)
	tflog.Debug(ctx, fmt.Sprintf("Will create group attribute assignment with id: %s", id))
	d.SetId(id)
	lc := m.(*LldapClient)
	addAttrErr := lc.AddAttributeToGroup(groupId, attributeId, value)
	if addAttrErr != nil {
		return addAttrErr
	}
	tflog.Info(ctx, fmt.Sprintf("Created group attribute assignment with id: %s", id))
	return nil
}

func resourceGroupAttributeAssignmentRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Id()
	groupIdString, attributeId, ok := strings.Cut(id, resourceGroupAttributeAssignmentIdSeparator)
	if !ok {
		return diag.Errorf("not a valid lldap_group_attribute_assignment id: %s", id)
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
	groupAttributes := make([]string, len(group.Attributes))
	var value []string
	for _, attr := range group.Attributes {
		groupAttributes = append(groupAttributes, attr.Name)
		if attr.Name == attributeId {
			value = attr.Value
		}
	}
	if !slices.Contains(groupAttributes, attributeId) {
		return diag.Errorf("Group is missing attribute!")
	}
	for k, v := range map[string]any{
		"group_id":     groupId,
		"attribute_id": attributeId,
		"value":        value,
	} {
		if setErr := d.Set(k, v); setErr != nil {
			return diag.FromErr(setErr)
		}
	}
	return nil
}

func resourceGroupAttributeAssignmentUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	groupId := d.Get("group_id").(int)
	attributeId := d.Get("attribute_id").(string)
	valueRaw := d.Get("value").(*schema.Set)
	valueRawList := valueRaw.List()
	value := make([]string, len(valueRawList))
	for i, vRaw := range valueRawList {
		value[i] = vRaw.(string)
	}
	lc := m.(*LldapClient)

	// First remove the existing attribute
	removeAttrErr := lc.RemoveAttributeFromGroup(groupId, attributeId)
	if removeAttrErr != nil {
		return removeAttrErr
	}

	// Then add it back with the new values
	updateAttrErr := lc.AddAttributeToGroup(groupId, attributeId, value)
	if updateAttrErr != nil {
		return updateAttrErr
	}
	tflog.Info(ctx, fmt.Sprintf("Updated group attribute assignment with id: %s", d.Id()))
	return nil
}

func resourceGroupAttributeAssignmentDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	groupId := d.Get("group_id").(int)
	attributeId := d.Get("attribute_id").(string)
	lc := m.(*LldapClient)
	removeAttrErr := lc.RemoveAttributeFromGroup(groupId, attributeId)
	if removeAttrErr != nil {
		return removeAttrErr
	}
	return nil
}
