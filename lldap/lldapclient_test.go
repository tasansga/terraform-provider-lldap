/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLldapUserGetCustomAttributes(t *testing.T) {
	user := LldapUser{
		Attributes: []LldapCustomAttribute{
			{
				Name: "first_name",
			},
			{
				Name: "custom",
			},
			{
				Name: "avatar",
			},
		},
	}
	expected := []LldapCustomAttribute{
		{
			Name: "custom",
		},
	}
	assert.Equal(t, expected, user.GetCustomAttributes())
}

func TestRemoveUserFromGroup_ErrorResponse(t *testing.T) {
	server, client := newMockLldapServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/simple/login" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token":        "mock-token",
				"refreshToken": "mock-refresh",
			})
			return
		}

		if r.URL.Path == "/api/graphql" {
			var query struct {
				OperationName string `json:"operationName"`
			}
			err := json.NewDecoder(r.Body).Decode(&query)
			require.NoError(t, err)

			if query.OperationName == "RemoveUserFromGroup" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"removeUserFromGroup": map[string]any{
							"ok": false,
						},
					},
				})
				return
			}
		}

		http.Error(w, "not found", http.StatusNotFound)
	})
	defer server.Close()

	diags := client.RemoveUserFromGroup(1, "alice")
	require.True(t, diags.HasError(), "expected error when removeUserFromGroup returns ok: false")
	assert.Contains(t, diags[0].Summary, "Failed to remove user from group")
}
