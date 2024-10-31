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
		"display_name":  group.DisplayName,
		"creation_date": group.CreationDate,
		"uuid":          group.Uuid,
		"users":         resourceGroupUsersParser(group.Users),
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
