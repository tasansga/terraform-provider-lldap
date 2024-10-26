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
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique username",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique user email",
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
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID of user",
			},
			"avatar": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Base 64 encoded JPEG image",
			},
			"groups": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Groups where the user is a member",
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
							Description: "Display name of the group",
						},
					},
				},
			},
		},
	}
}

func dataSourceUserMembershipParser(groups []LldapGroup) []map[string]any {
	result := make([]map[string]any, len(groups))
	for i, llgroup := range groups {
		group := map[string]any{
			"id":           llgroup.Id,
			"display_name": llgroup.DisplayName,
		}
		result[i] = group
	}
	return result
}

func dataSourceUserRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Get("id").(string)
	lc := m.(*LldapClient)
	user, getUserErr := lc.GetUser(id)
	if getUserErr != nil {
		return getUserErr
	}
	d.SetId(user.Id)
	for k, v := range map[string]interface{}{
		"username":      user.Id,
		"email":         user.Email,
		"display_name":  user.DisplayName,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"creation_date": user.CreationDate,
		"uuid":          user.Uuid,
		"avatar":        user.Avatar,
		"groups":        dataSourceUserMembershipParser(user.Groups),
	} {
		if setErr := d.Set(k, v); setErr != nil {
			return diag.FromErr(setErr)
		}
	}
	return nil
}
