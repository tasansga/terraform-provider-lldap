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

func TestResourceGroupMembershipsImport(t *testing.T) {
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

			if query.OperationName == "GetGroupDetails" {
				var vars struct {
					Id int `json:"id"`
				}
				_ = json.Unmarshal(query.Variables, &vars)

				if vars.Id == 4 {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{
						"data": map[string]any{
							"group": map[string]any{
								"id":          4,
								"displayName": "serviceaccounts",
								"users": []map[string]any{
									{"id": "svc-authelia"},
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

	r := resourceGroupMemberships()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	d.SetId("4")

	importer := r.Importer
	require.NotNil(t, importer)

	importedData, importErr := importer.StateContext(context.Background(), d, client)
	require.NoError(t, importErr)
	require.Len(t, importedData, 1)

	readDiags := resourceGroupMembershipsRead(context.Background(), importedData[0], client)
	require.False(t, readDiags.HasError(), "read failed: %v", readDiags)

	assert.Equal(t, "4", importedData[0].Get("group_id"))
	assert.Equal(t, "4", importedData[0].Id())

	rawUserIds := importedData[0].Get("user_ids").(*schema.Set).List()
	assert.ElementsMatch(t, []any{"svc-authelia"}, rawUserIds)
}

func TestResourceGroupMembershipsImport_InvalidId(t *testing.T) {
	r := resourceGroupMemberships()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	d.SetId("not-a-number")

	importer := r.Importer
	require.NotNil(t, importer)

	_, err := importer.StateContext(context.Background(), d, nil)
	require.Error(t, err, "expected error when importing non-integer group id")
	assert.Contains(t, err.Error(), "not a valid group_id")
}

func TestResourceGroupMembershipsRead_NotFound(t *testing.T) {
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
						"message": "Entity not found: `99`",
						"path":    []string{"group"},
					},
				},
			})
			return
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	r := resourceGroupMemberships()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id": "99",
	})
	d.SetId("99")

	readDiags := resourceGroupMembershipsRead(context.Background(), d, client)
	require.False(t, readDiags.HasError(), "expected no error for deleted group, got %v", readDiags)
	assert.Empty(t, d.Id(), "expected id to be cleared when group is not found")
}

func TestResourceGroupMembershipsRead_InvalidGroupId(t *testing.T) {
	r := resourceGroupMemberships()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id": "invalid-int",
	})
	d.SetId("invalid-int")

	readDiags := resourceGroupMembershipsRead(context.Background(), d, nil)
	require.True(t, readDiags.HasError(), "expected error for non-integer group_id")
}

func TestResourceGroupMembershipsCreate_InvalidGroupId(t *testing.T) {
	r := resourceGroupMemberships()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id": "invalid-int",
		"user_ids": []any{},
	})

	createDiags := resourceGroupMembershipsCreate(context.Background(), d, nil)
	require.True(t, createDiags.HasError(), "expected error for non-integer group_id")
}

func TestResourceGroupMembershipsUpdate_InvalidGroupId(t *testing.T) {
	r := resourceGroupMemberships()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id": "invalid-int",
		"user_ids": []any{},
	})
	d.SetId("invalid-int")

	updateDiags := resourceGroupMembershipsUpdate(context.Background(), d, nil)
	require.True(t, updateDiags.HasError(), "expected error for non-integer group_id")
}

func TestResourceGroupMembershipsDelete_InvalidGroupId(t *testing.T) {
	r := resourceGroupMemberships()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id": "invalid-int",
		"user_ids": []any{},
	})

	deleteDiags := resourceGroupMembershipsDelete(context.Background(), d, nil)
	require.True(t, deleteDiags.HasError(), "expected error for non-integer group_id")
}

func TestResourceGroupMemberships_Description(t *testing.T) {
	r := resourceGroupMemberships()
	assert.Equal(t, "Exclusively manages all LLDAP memberships for this specific group", r.Description)
}


