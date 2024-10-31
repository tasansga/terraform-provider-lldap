package lldap

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupsRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"groups": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of all groups",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The unique group ID",
						},
						"display_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Display name of this group",
						},
						"creation_date": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Metadata of group object creation",
						},
					},
				},
			},
		},
	}
}

func dataSourceGroupsRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lc := m.(*LldapClient)
	llgroups, getGroupsErr := lc.GetGroups()
	if getGroupsErr != nil {
		return getGroupsErr
	}
	hashBase, marshalErr := json.Marshal(llgroups)
	if marshalErr != nil {
		return diag.FromErr(marshalErr)
	}
	hash := sha1.New()
	hash.Write([]byte(hashBase))
	hashString := hex.EncodeToString(hash.Sum(nil))
	d.SetId(hashString)
	if setErr := d.Set("groups", dataSourceGroupsParser(llgroups)); setErr != nil {
		return diag.Errorf("could not create group set: %s", setErr)
	}
	return nil
}

func dataSourceGroupsParser(llgroups []LldapGroup) []map[string]any {
	result := make([]map[string]any, len(llgroups))
	for i, llgroup := range llgroups {
		group := map[string]any{
			"id":           strconv.Itoa(llgroup.Id),
			"display_name": llgroup.DisplayName,
		}
		result[i] = group
	}
	return result
}
