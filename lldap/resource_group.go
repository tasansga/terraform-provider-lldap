package lldap

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_computed_optional": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"not_computed_required": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceGroupCreate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	fmt.Printf("ResourceData: %+v\n", d)
	return nil
}

func resourceGroupRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	fmt.Printf("ResourceData: %+v\n", d)
	return nil
}

func resourceGroupUpdate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	fmt.Printf("ResourceData: %+v\n", d)
	return nil
}

func resourceGroupDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	fmt.Printf("ResourceData: %+v\n", d)
	return nil
}
