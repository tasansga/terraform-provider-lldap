/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package lldap

import "strconv"

func dataSourceAttributesParser(attrs []LldapCustomAttribute) []map[string]any {
	result := make([]map[string]any, len(attrs))
	for i, llattr := range attrs {
		attr := map[string]any{
			"name":  llattr.Name,
			"value": llattr.Value,
		}
		result[i] = attr
	}
	return result
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
