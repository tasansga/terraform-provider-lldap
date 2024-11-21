/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package lldap

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSetHashId(d *schema.ResourceData, v any) diag.Diagnostics {
	hashBase, marshalErr := json.Marshal(v)
	if marshalErr != nil {
		return diag.FromErr(marshalErr)
	}
	hash := sha1.New()
	hash.Write([]byte(hashBase))
	hashString := hex.EncodeToString(hash.Sum(nil))
	d.SetId(hashString)
	return nil
}

func dataSourceGroupsParser(llgroups []LldapGroup) []map[string]any {
	result := make([]map[string]any, len(llgroups))
	for i, llgroup := range llgroups {
		group := map[string]any{
			"id":            strconv.Itoa(llgroup.Id),
			"display_name":  llgroup.DisplayName,
			"creation_date": llgroup.CreationDate,
		}
		result[i] = group
	}
	return result
}
