/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
				_ = d.Set("id", d.Id())
				return schema.ImportStatePassthroughContext(ctx, d, m)
			},
		},
		Description: "Manages a LLDAP user",
		Schema: map[string]*schema.Schema{
			"attributes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Attributes for this user",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name of this attribute",
						},
						"value": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of values for this attribute",
						},
					},
				},
			},
			"avatar": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Base 64 encoded JPEG image",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Metadata of user object creation",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Display name of this user",
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique user email",
			},
			"first_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "First name of this user",
			},
			"groups": &dataSourceGroupsSchema,
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID representing this specific user",
			},
			"last_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Last name of this user",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password for the user. Note that the provider cannot read the password from LLDAP, so if this value is not set, the password attribute will be entirely ignored by the provider.",
			},
			"password_wo": {
				Type:          schema.TypeString,
				Optional:      true,
				WriteOnly:     true,
				Description:   "Password for the user as a write-only attribute. Increment password_wo_version when this value needs to be updated.",
				ConflictsWith: []string{"password"},
				RequiredWith:  []string{"password_wo_version"},
			},
			"password_wo_version": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Triggers an update of password_wo. Increment this value when a password update is needed.",
				ConflictsWith: []string{"password"},
				RequiredWith:  []string{"password_wo"},
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique username",
				StateFunc: func(val any) string {
					return strings.ToLower(val.(string))
				},
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID of user",
			},
		},
		ValidateRawResourceConfigFuncs: []schema.ValidateRawResourceConfigFunc{
			validation.PreferWriteOnlyAttribute(cty.GetAttrPath("password"), cty.GetAttrPath("password_wo")),
		},
	}
}

func resourceUserGetWriteOnlyPassword(d *schema.ResourceData) (string, diag.Diagnostics) {
	value, diags := d.GetRawConfigAt(cty.GetAttrPath("password_wo"))

	if diags.HasError() {
		return "", diags
	}

	if !value.IsKnown() || value.IsNull() {
		return "", nil
	}

	return value.AsString(), nil
}

func resourceUserSetResourceData(d *schema.ResourceData, user *LldapUser) diag.Diagnostics {
	for k, v := range map[string]any{
		"attributes":    attributesParser(user.Attributes),
		"avatar":        user.Avatar,
		"creation_date": user.CreationDate,
		"display_name":  user.DisplayName,
		"email":         user.Email,
		"first_name":    user.FirstName,
		"groups":        dataSourceGroupsParser(user.Groups),
		"last_name":     user.LastName,
		"username":      user.Id,
		"uuid":          user.Uuid,
	} {
		if setErr := d.Set(k, v); setErr != nil {
			return diag.FromErr(setErr)
		}
	}

	return nil
}

func resourceUserGetResourceData(d *schema.ResourceData) LldapUser {
	return LldapUser{
		Id:          d.Get("username").(string),
		Email:       d.Get("email").(string),
		DisplayName: d.Get("display_name").(string),
		FirstName:   d.Get("first_name").(string),
		LastName:    d.Get("last_name").(string),
		Uuid:        d.Get("uuid").(string),
		Avatar:      d.Get("avatar").(string),
	}
}

func resourceUserCreate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	user := resourceUserGetResourceData(d)

	var password string
	var isWriteOnlyPassword bool

	if _, ok := d.GetOk("password"); !ok {
		writeOnlyPassword, diags := resourceUserGetWriteOnlyPassword(d)

		if diags.HasError() {
			return diags
		}

		password = writeOnlyPassword
		isWriteOnlyPassword = true
	} else {
		password = d.Get("password").(string)
		isWriteOnlyPassword = false
	}

	lc := m.(*LldapClient)
	createErr := lc.CreateUser(&user)
	if createErr != nil {
		return createErr
	}
	if password != "" {
		setPwErr := lc.SetUserPassword(user.Id, password)
		if setPwErr != nil {
			return setPwErr
		}
	}
	d.SetId(user.Id)

	setRdErr := resourceUserSetResourceData(d, &user)
	if setRdErr != nil {
		return setRdErr
	}

	if !isWriteOnlyPassword {
		if setErr := d.Set("password", password); setErr != nil {
			return diag.FromErr(setErr)
		}
	}

	return nil
}

func resourceUserRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	user, getUserErr := lc.GetUser(d.Id())
	if getUserErr != nil {
		// If the user was not found, mark the resource as deleted so Terraform will recreate it
		if isEntityNotFoundError(getUserErr) {
			d.SetId("")
			return nil
		}
		return getUserErr
	}

	setRdErr := resourceUserSetResourceData(d, user)
	if setRdErr != nil {
		return setRdErr
	}

	// We cannot read the password from LLDAP, but we can check whether the value from state is still valid.
	statePassword := d.Get("password").(string)
	if statePassword != "" {
		isValidPassword, _ := lc.IsValidPassword(user.Id, statePassword)

		if isValidPassword {
			if setErr := d.Set("password", statePassword); setErr != nil {
				return diag.FromErr(setErr)
			}
		}
	}

	return nil
}

func resourceUserUpdate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	user := resourceUserGetResourceData(d)
	updateErr := lc.UpdateUser(&user)
	if updateErr != nil {
		return updateErr
	}

	if d.HasChange("password_wo_version") {
		password, diags := resourceUserGetWriteOnlyPassword(d)
		if diags.HasError() {
			return diags
		}

		if password == "" {
			return diag.Errorf("password_wo must be non-empty when password_wo_version changes")
		}

		if setPwErr := lc.SetUserPassword(user.Id, password); setPwErr != nil {
			return setPwErr
		}

		return nil
	}

	statePassword := d.Get("password").(string)

	if statePassword != "" {
		isValidPassword, bindErr := lc.IsValidPassword(user.Id, statePassword)
		if bindErr != nil {
			return bindErr
		}
		if !isValidPassword {
			setPwErr := lc.SetUserPassword(user.Id, statePassword)
			if setPwErr != nil {
				return setPwErr
			}
		}
	}
	return nil
}

func resourceUserDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	deleteErr := lc.DeleteUser(d.Id())
	if deleteErr != nil {
		return deleteErr
	}
	return nil
}
