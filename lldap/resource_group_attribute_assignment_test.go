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

func TestResourceGroupAttributeAssignmentRead_GroupNotFound(t *testing.T) {
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

	r := resourceGroupAttributeAssignment()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id":     99,
		"attribute_id": "attr1",
	})
	d.SetId("99:attr1")

	readDiags := resourceGroupAttributeAssignmentRead(context.Background(), d, client)
	require.False(t, readDiags.HasError(), "expected no error for deleted group, got %v", readDiags)
	assert.Empty(t, d.Id(), "expected id to be cleared when group is not found")
}

func TestResourceGroupAttributeAssignmentRead_Success(t *testing.T) {
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
				"data": map[string]any{
					"group": map[string]any{
						"id":          1,
						"displayName": "developers",
						"attributes": []map[string]any{
							{
								"name":  "custom_attr",
								"value": []string{"val1", "val2"},
							},
						},
					},
				},
			})
			return
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	r := resourceGroupAttributeAssignment()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id":     1,
		"attribute_id": "custom_attr",
	})
	d.SetId("1:custom_attr")

	readDiags := resourceGroupAttributeAssignmentRead(context.Background(), d, client)
	require.False(t, readDiags.HasError(), "expected no error, got %v", readDiags)
	assert.Equal(t, "1:custom_attr", d.Id())
	assert.Equal(t, 1, d.Get("group_id"))
	assert.Equal(t, "custom_attr", d.Get("attribute_id"))
	rawValues := d.Get("value").(*schema.Set).List()
	assert.ElementsMatch(t, []any{"val1", "val2"}, rawValues)
}

func TestResourceGroupAttributeAssignmentRead_AttributeMissing(t *testing.T) {
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
				"data": map[string]any{
					"group": map[string]any{
						"id":          1,
						"displayName": "developers",
						"attributes":  []map[string]any{},
					},
				},
			})
			return
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	r := resourceGroupAttributeAssignment()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id":     1,
		"attribute_id": "custom_attr",
	})
	d.SetId("1:custom_attr")

	readDiags := resourceGroupAttributeAssignmentRead(context.Background(), d, client)
	require.False(t, readDiags.HasError(), "expected no error, got %v", readDiags)
	assert.Empty(t, d.Id(), "expected id to be cleared when attribute is missing remotely")
}

func TestResourceGroupAttributeAssignmentRead_InvalidId(t *testing.T) {
	r := resourceGroupAttributeAssignment()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	d.SetId("invalid-id-without-separator")

	readDiags := resourceGroupAttributeAssignmentRead(context.Background(), d, nil)
	require.True(t, readDiags.HasError())
	assert.Contains(t, readDiags[0].Summary, "not a valid lldap_group_attribute_assignment id")
}

