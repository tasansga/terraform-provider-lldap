package lldap

import (
	"context"
	"fmt"
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
				ForceNew:    true,
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
	displayName := d.Get("display_name").(string)
	group := LldapGroup{
		DisplayName: displayName,
	}
	lc := m.(*LldapClient)
	createErr := lc.CreateGroup(&group)
	if createErr != nil {
		return createErr
	}
	d.SetId(strconv.Itoa(group.Id))
	if setErr := d.Set("display_name", group.DisplayName); setErr != nil {
		return diag.FromErr(setErr)
	}
	if group.CreationDate != nil {
		if setErr := d.Set("creation_date", group.CreationDate); setErr != nil {
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
	//lc := m.(*LldapClient)
	fmt.Printf("ResourceData: %+v\n", d)
	return diag.Errorf("Not implemented: resourceGroupUpdate")
}

func resourceGroupDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	//lc := m.(*LldapClient)
	fmt.Printf("ResourceData: %+v\n", d)
	return diag.Errorf("Not implemented: resourceGroupDelete")
}
