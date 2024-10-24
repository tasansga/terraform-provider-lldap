package lldap

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				_ = d.Set("id", d.Id())
				return schema.ImportStatePassthroughContext(ctx, d, m)
			},
		},
		DeleteContext: resourceUserDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID representing this specific user with email",
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
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique user email",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Display name of this user",
			},
			"first_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "First name of this user",
			},
			"last_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Last name of this user",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Metadata of user object creation",
			},
		},
	}
}

func resourceUserCreate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	user := LldapUser{
		Id:          d.Get("username").(string),
		Email:       d.Get("email").(string),
		DisplayName: d.Get("display_name").(string),
		FirstName:   d.Get("first_name").(string),
		LastName:    d.Get("last_name").(string),
	}
	lc := m.(*LldapClient)
	createErr := lc.CreateUser(&user)
	if createErr != nil {
		return createErr
	}
	d.SetId(user.Id)
	for k, v := range map[string]string{
		"username":      user.Id,
		"email":         user.Email,
		"display_name":  user.DisplayName,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"creation_date": user.CreationDate,
	} {
		if setErr := d.Set(k, v); setErr != nil {
			return diag.FromErr(setErr)
		}
	}
	return nil
}

func resourceUserRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	//lc := m.(*LldapClient)
	fmt.Printf("ResourceData: %+v\n", d)
	return nil
}

func resourceUserUpdate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	//lc := m.(*LldapClient)
	fmt.Printf("ResourceData: %+v\n", d)
	return nil
}

func resourceUserDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	//lc := m.(*LldapClient)
	fmt.Printf("ResourceData: %+v\n", d)
	return nil
}
