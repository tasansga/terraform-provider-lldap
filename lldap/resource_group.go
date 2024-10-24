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
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique group ID",
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Display name of this group",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Metadata of group object creation",
			},
			"users": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of users who are members of this group",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
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
	for k, v := range map[string]string{
		"display_name":  group.DisplayName,
		"creation_date": *group.CreationDate,
	} {
		if setErr := d.Set(k, v); setErr != nil {
			return diag.FromErr(setErr)
		}
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
	if setErr := d.Set("display_name", group.DisplayName); setErr != nil {
		return diag.FromErr(setErr)
	}
	if setErr := d.Set("creation_date", group.CreationDate); setErr != nil {
		return diag.FromErr(setErr)
	}
	if setErr := d.Set("users", resourceGroupUsersParser(group.Users)); setErr != nil {
		return diag.FromErr(setErr)
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
