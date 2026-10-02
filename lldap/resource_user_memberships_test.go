/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceUserMembershipsImport(t *testing.T) {
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
				OperationName string          `json:"operationName"`
				Variables     json.RawMessage `json:"variables"`
			}
			err := json.NewDecoder(r.Body).Decode(&query)
			require.NoError(t, err)

			if query.OperationName == "GetUserDetails" {
				var vars struct {
					Id string `json:"id"`
				}
				_ = json.Unmarshal(query.Variables, &vars)

				if vars.Id == "test-user" {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{
						"data": map[string]any{
							"user": map[string]any{
								"id":          "test-user",
								"email":       "test@example.com",
								"displayName": "Test User",
								"groups": []map[string]any{
									{"id": 2, "displayName": "lldap_password_manager"},
								},
							},
						},
					})
					return
				}
			}
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	r := resourceUserMemberships()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	d.SetId("test-user")

	importer := r.Importer
	require.NotNil(t, importer)

	importedData, importErr := importer.StateContext(context.Background(), d, client)
	require.NoError(t, importErr)
	require.Len(t, importedData, 1)

	readDiags := resourceUserMembershipsRead(context.Background(), importedData[0], client)
	require.False(t, readDiags.HasError(), "read failed: %v", readDiags)

	assert.Equal(t, "test-user", importedData[0].Get("user_id"))
	assert.Equal(t, "test-user", importedData[0].Id())

	rawGroupIds := importedData[0].Get("group_ids").(*schema.Set).List()
	assert.ElementsMatch(t, []any{"2"}, rawGroupIds)
}

func TestResourceUserMembershipsRead_NotFound(t *testing.T) {
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
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": nil,
				"errors": []map[string]any{
					{
						"message": "Entity not found: `deleted-user`",
						"path":    []string{"user"},
					},
				},
			})
			return
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	r := resourceUserMemberships()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"user_id": "deleted-user",
	})
	d.SetId("deleted-user")

	readDiags := resourceUserMembershipsRead(context.Background(), d, client)
	require.False(t, readDiags.HasError(), "expected no error for deleted user, got %v", readDiags)
	assert.Empty(t, d.Id(), "expected id to be cleared when user is not found")
}

func TestResourceUserMemberships_Description(t *testing.T) {
	r := resourceUserMemberships()
	assert.Equal(t, "Exclusively manages all LLDAP memberships for this specific user", r.Description)
}

