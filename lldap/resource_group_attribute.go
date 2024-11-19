/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGroupAttribute() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupAttributeCreate,
		ReadContext:   resourceGroupAttributeRead,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				_ = d.Set("id", d.Id())
				return schema.ImportStatePassthroughContext(ctx, d, m)
			},
		},
		DeleteContext: resourceGroupAttributeDelete,
		Schema: map[string]*schema.Schema{
			"attribute_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("The attribute type (one of: %s)", strings.Join(VALID_ATTRIBUTE_TYPES[:], ", ")),
				ValidateDiagFunc: func(v any, p cty.Path) diag.Diagnostics {
					attributeType := v.(string)
					_, parseErr := parseLldapCustomAttributeType(attributeType)
					if parseErr != nil {
						return diag.FromErr(parseErr)
					}
					return nil
				},
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The attribute type",
			},
			"is_list": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Does this represent a list?",
			},
			"is_visible": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     true,
				Description: "Is this attribute visible in LDAP?",
			},
			"is_readonly": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Is this attribute readonly (i.e. managed by LLDAP)?",
			},
			"is_hardcoded": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Is this attribute hardcoded (i.e. managed by LLDAP)?",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique attribute name",
			},
		},
	}
}

func resourceGroupAttributeSetResourceData(d *schema.ResourceData, schema *LldapGroupAttributeSchema) diag.Diagnostics {
	d.SetId(schema.Name)
	for k, v := range map[string]interface{}{
		"name":           schema.Name,
		"attribute_type": schema.AttributeType,
		"is_list":        schema.IsList,
		"is_visible":     schema.IsVisible,
		"is_readonly":    false,
		"is_hardcoded":   false,
	} {
		if setErr := d.Set(k, v); setErr != nil {
			return diag.FromErr(setErr)
		}
	}
	return nil
}

func resourceGroupAttributeGetResourceData(d *schema.ResourceData) (*LldapGroupAttributeSchema, error) {
	attributeType, attrTypeErr := parseLldapCustomAttributeType(d.Get("attribute_type").(string))
	if attrTypeErr != nil {
		return nil, attrTypeErr
	}
	return &LldapGroupAttributeSchema{
		Name:          d.Get("name").(string),
		AttributeType: *attributeType,
		IsList:        d.Get("is_list").(bool),
		IsVisible:     d.Get("is_visible").(bool),
		IsHardcoded:   false,
		IsReadonly:    false,
	}, nil
}

func resourceGroupAttributeCreate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	schema, getAttrErr := resourceGroupAttributeGetResourceData(d)
	if getAttrErr != nil {
		return diag.FromErr(getAttrErr)
	}
	createAttrErr := lc.CreateGroupAttribute(schema.Name, schema.AttributeType, schema.IsList, schema.IsVisible)
	if createAttrErr != nil {
		return createAttrErr
	}
	createdSchema, getSchemaErr := lc.GetGroupAttributeSchema(schema.Name)
	if getSchemaErr != nil {
		return getSchemaErr
	}
	setRdErr := resourceGroupAttributeSetResourceData(d, createdSchema)
	if setRdErr != nil {
		return setRdErr
	}
	return nil
}

func resourceGroupAttributeRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	schema, getSchemaErr := lc.GetGroupAttributeSchema(d.Get("name").(string))
	if getSchemaErr != nil {
		return getSchemaErr
	}
	setRdErr := resourceGroupAttributeSetResourceData(d, schema)
	if setRdErr != nil {
		return setRdErr
	}
	return nil
}

func resourceGroupAttributeDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	deleteErr := lc.DeleteGroupAttribute(d.Id())
	if deleteErr != nil {
		return deleteErr
	}
	return nil
}
