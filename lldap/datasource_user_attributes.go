/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUserAttributes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserAttributesRead,
		Schema: map[string]*schema.Schema{
			"attributes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of all user attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique attribute name",
						},
						"attribute_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The attribute type",
						},
						"is_list": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Does this represent a list?",
						},
						"is_visible": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Is this attribute visible in LDAP?",
						},
						"is_editable": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Is this attribute user editable?",
						},
						"is_hardcoded": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Is this attribute hardcoded (i.e. managed by LLDAP)?",
						},
						"is_readonly": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Is this attribute readonly (i.e. managed by LLDAP)?",
						},
					},
				},
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Generated ID representing the attributes",
			},
		},
	}
}

func dataSourceUserAttributeSchemaParser(schemas []LldapUserAttributeSchema) []map[string]any {
	result := make([]map[string]any, len(schemas))
	for i, llattr := range schemas {
		attr := map[string]any{
			"name":           llattr.Name,
			"attribute_type": llattr.AttributeType,
			"is_list":        llattr.IsList,
			"is_visible":     llattr.IsVisible,
			"is_editable":    llattr.IsEditable,
			"is_hardcoded":   llattr.IsHardcoded,
			"is_readonly":    llattr.IsReadonly,
		}
		result[i] = attr
	}
	return result
}

func dataSourceUserAttributesRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	schemas, getSchemaErr := lc.GetUserAttributesSchema()
	if getSchemaErr != nil {
		return getSchemaErr
	}
	dataSourceSetHashId(d, schemas)
	if setErr := d.Set("attributes", dataSourceUserAttributeSchemaParser(schemas)); setErr != nil {
		return diag.Errorf("could not create attribute schema set: %s", setErr)
	}
	return nil
}
