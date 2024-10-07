package lldap

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMemberCreate,
		ReadContext:   resourceMemberRead,
		UpdateContext: resourceMemberUpdate,
		DeleteContext: resourceMemberDelete,
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
				Description: "TODO: The member 'ID', something like group_id:user_id or whatever?",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique group ID",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique user ID",
			},
		},
	}
}

func resourceMemberCreate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	//lc := m.(*LldapClient)
	fmt.Printf("ResourceData: %+v\n", d)
	return diag.Errorf("Not implemented: resourceGroupUpdate")
}

func resourceMemberRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	//lc := m.(*LldapClient)
	fmt.Printf("ResourceData: %+v\n", d)
	return diag.Errorf("Not implemented: resourceGroupUpdate")
}

func resourceMemberUpdate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	//lc := m.(*LldapClient)
	fmt.Printf("ResourceData: %+v\n", d)
	return diag.Errorf("Not implemented: resourceGroupUpdate")
}

func resourceMemberDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	//lc := m.(*LldapClient)
	fmt.Printf("ResourceData: %+v\n", d)
	return diag.Errorf("Not implemented: resourceGroupDelete")
}
