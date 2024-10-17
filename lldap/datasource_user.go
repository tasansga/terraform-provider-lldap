package lldap

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique user ID",
				StateFunc: func(val any) string {
					return strings.ToLower(val.(string))
				},
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique user ID",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Display name of this user",
			},
			"first_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "First name of this user",
			},
			"last_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last name of this user",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Metadata of user object creation",
			},
			"groups": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Members of this group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The unique group ID",
						},
						"display_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Display name of this group",
						},
					},
				},
			},
		},
	}
}

func dataSourceUserRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Get("id").(string)
	lc := m.(*LldapClient)
	lluser, getUserErr := lc.GetUser(id)
	if getUserErr != nil {
		return getUserErr
	}
	d.SetId(lluser.Id)
	d.Set("email", lluser.Email)
	if setErr := d.Set("display_name", lluser.DisplayName); setErr != nil {
		return diag.Errorf("Could not set display_name: %s", setErr)
	}
	if setErr := d.Set("first_name", lluser.FirstName); setErr != nil {
		return diag.Errorf("Could not set first_name: %s", setErr)
	}
	if setErr := d.Set("last_name", lluser.LastName); setErr != nil {
		return diag.Errorf("Could not set last_name: %s", setErr)
	}
	if setErr := d.Set("creation_date", lluser.CreationDate); setErr != nil {
		return diag.Errorf("Could not set creation_date: %s", setErr)
	}
	if setErr := d.Set("groups", dataSourceUserMembershipParser(lluser.Groups)); setErr != nil {
		return diag.Errorf("Could not set groups: %s", setErr)
	}
	return nil
}

func dataSourceUserMembershipParser(llgroups []LldapGroup) []map[string]any {
	result := make([]map[string]any, len(llgroups))
	for i, llgroup := range llgroups {
		group := map[string]any{
			"id":           llgroup.Id,
			"display_name": llgroup.DisplayName,
		}
		result[i] = group
	}
	return result
}
