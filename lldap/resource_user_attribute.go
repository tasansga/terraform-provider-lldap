/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUserAttribute() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserAttributeCreate,
		ReadContext:   resourceUserAttributeRead,
		DeleteContext: resourceUserAttributeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
				_ = d.Set("id", d.Id())
				return schema.ImportStatePassthroughContext(ctx, d, m)
			},
		},
		Description: "Defines a new custom attribute schema for users",
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
			"is_editable": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Is this attribute user editable?",
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

func resourceUserAttributeSetResourceData(d *schema.ResourceData, schema *LldapUserAttributeSchema) diag.Diagnostics {
	d.SetId(schema.Name)
	for k, v := range map[string]any{
		"name":           schema.Name,
		"attribute_type": schema.AttributeType,
		"is_list":        schema.IsList,
		"is_visible":     schema.IsVisible,
		"is_editable":    schema.IsEditable,
		"is_readonly":    false,
		"is_hardcoded":   false,
	} {
		if setErr := d.Set(k, v); setErr != nil {
			return diag.FromErr(setErr)
		}
	}
	return nil
}

func parseLldapCustomAttributeType(value string) (*LldapCustomAttributeType, error) {
	if !slices.Contains(VALID_ATTRIBUTE_TYPES, value) {
		return nil, fmt.Errorf("invalid value for attribute type: '%s' Valid values are: %s", value, strings.Join(VALID_ATTRIBUTE_TYPES[:], ", "))
	}
	result := LldapCustomAttributeType(value)
	return &result, nil
}

func resourceUserAttributeGetResourceData(d *schema.ResourceData) (*LldapUserAttributeSchema, error) {
	attributeType, attrTypeErr := parseLldapCustomAttributeType(d.Get("attribute_type").(string))
	if attrTypeErr != nil {
		return nil, attrTypeErr
	}
	return &LldapUserAttributeSchema{
		Name:          d.Get("name").(string),
		AttributeType: *attributeType,
		IsList:        d.Get("is_list").(bool),
		IsVisible:     d.Get("is_visible").(bool),
		IsEditable:    d.Get("is_editable").(bool),
		IsHardcoded:   false,
		IsReadonly:    false,
	}, nil
}

func resourceUserAttributeCreate(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	schema, getAttrErr := resourceUserAttributeGetResourceData(d)
	if getAttrErr != nil {
		return diag.FromErr(getAttrErr)
	}
	createAttrErr := lc.CreateUserAttribute(schema.Name, schema.AttributeType, schema.IsList, schema.IsVisible, schema.IsEditable)
	if createAttrErr != nil {
		return createAttrErr
	}
	createdSchema, getSchemaErr := lc.GetUserAttributeSchema(schema.Name)
	if getSchemaErr != nil {
		return getSchemaErr
	}
	setRdErr := resourceUserAttributeSetResourceData(d, createdSchema)
	if setRdErr != nil {
		return setRdErr
	}
	return nil
}

func resourceUserAttributeRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	schema, getSchemaErr := lc.GetUserAttributeSchema(d.Id())
	if getSchemaErr != nil {
		return getSchemaErr
	}
	setRdErr := resourceUserAttributeSetResourceData(d, schema)
	if setRdErr != nil {
		return setRdErr
	}
	return nil
}

func resourceUserAttributeDelete(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	deleteErr := lc.DeleteUserAttribute(d.Id())
	if deleteErr != nil {
		return deleteErr
	}
	return nil
}
