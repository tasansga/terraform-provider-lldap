package lldap

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The unique group ID",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Display name of this group",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Metadata of group object creation",
			},
			"users": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Members of this group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique user ID",
						},
						"display_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Display name of this user",
						},
					},
				},
			},
		},
	}
}

func dataSourceGroupRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Get("id").(int)
	lc := m.(*LldapClient)
	llgroup, getGroupErr := lc.GetGroup(id)
	if getGroupErr != nil {
		return getGroupErr
	}
	d.SetId(strconv.Itoa(llgroup.Id))
	d.Set("display_name", llgroup.DisplayName)
	d.Set("creation_date", llgroup.CreationDate)
	d.Set("users", dataSourceGroupUsersParser(llgroup.Users))
	return nil
}

func dataSourceGroupUsersParser(llusers []LldapUser) []map[string]any {
	result := make([]map[string]any, len(llusers))
	for i, lluser := range llusers {
		group := map[string]any{
			"id":           lluser.Id,
			"display_name": lluser.DisplayName,
		}
		result[i] = group
	}
	return result
}
