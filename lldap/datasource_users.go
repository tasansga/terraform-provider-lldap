package lldap

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			"users": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of all users",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique user ID",
						},
						"email": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique user ID",
						},
						"display_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Display name of this user",
						},
						"first_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "First name of this user",
						},
						"last_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Last name of this user",
						},
						"creation_date": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Metadata of user object creation",
						},
					},
				},
			},
		},
	}
}

func dataSourceUsersRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	llusers, getUsersErr := lc.GetUsers()
	if getUsersErr != nil {
		return getUsersErr
	}
	hashBase, marshalErr := json.Marshal(llusers)
	if marshalErr != nil {
		return diag.FromErr(marshalErr)
	}
	hash := sha1.New()
	hash.Write([]byte(hashBase))
	hashString := hex.EncodeToString(hash.Sum(nil))
	d.SetId(hashString)
	if setErr := d.Set("users", dataSourceUsersParser(llusers)); setErr != nil {
		return diag.Errorf("could not create user set: %s", setErr)
	}
	return nil
}

func dataSourceUsersParser(llusers []LldapUser) []map[string]any {
	result := make([]map[string]any, len(llusers))
	for i, lluser := range llusers {
		user := map[string]any{
			"id":            lluser.Id,
			"email":         lluser.Email,
			"display_name":  lluser.DisplayName,
			"first_name":    lluser.FirstName,
			"last_name":     lluser.LastName,
			"creation_date": lluser.CreationDate,
		}
		result[i] = user
	}
	return result
}
