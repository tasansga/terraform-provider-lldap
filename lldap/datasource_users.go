/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

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
						"avatar": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Base 64 encoded JPEG image",
						},
						"creation_date": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Metadata of user object creation",
						},
						"display_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Display name of this user",
						},
						"email": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique user ID",
						},
						"first_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "First name of this user",
						},
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique user ID",
						},
						"last_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Last name of this user",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique username",
						},
						"uuid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "UUID of user",
						},
					},
				},
			},
		},
	}
}

func dataSourceUsersRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	users, getUsersErr := lc.GetUsers()
	if getUsersErr != nil {
		return getUsersErr
	}
	hashBase, marshalErr := json.Marshal(users)
	if marshalErr != nil {
		return diag.FromErr(marshalErr)
	}
	hash := sha1.New()
	hash.Write([]byte(hashBase))
	hashString := hex.EncodeToString(hash.Sum(nil))
	d.SetId(hashString)
	if setErr := d.Set("users", dataSourceUsersParser(users)); setErr != nil {
		return diag.Errorf("could not create user set: %s", setErr)
	}
	return nil
}

func dataSourceUsersParser(users []LldapUser) []map[string]any {
	result := make([]map[string]any, len(users))
	for i, user := range users {
		user := map[string]any{
			"id":            user.Id,
			"username":      user.Id,
			"email":         user.Email,
			"display_name":  user.DisplayName,
			"first_name":    user.FirstName,
			"last_name":     user.LastName,
			"creation_date": user.CreationDate,
			"uuid":          user.Uuid,
			"avatar":        user.Avatar,
		}
		result[i] = user
	}
	return result
}
