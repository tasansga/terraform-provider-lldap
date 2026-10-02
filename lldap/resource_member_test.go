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

func TestResourceMemberRead_GroupNotFound(t *testing.T) {
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

	r := resourceMember()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id": 99,
		"user_id":  "alice",
	})
	d.SetId("99:alice")

	readDiags := resourceMemberRead(context.Background(), d, client)
	require.False(t, readDiags.HasError(), "expected no error for deleted group, got %v", readDiags)
	assert.Empty(t, d.Id(), "expected id to be cleared when group is not found")
}

func TestResourceMemberRead_MembershipPresent(t *testing.T) {
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
						"users": []map[string]any{
							{"id": "alice"},
						},
					},
				},
			})
			return
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	r := resourceMember()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id": 1,
		"user_id":  "alice",
	})
	d.SetId("1:alice")

	readDiags := resourceMemberRead(context.Background(), d, client)
	require.False(t, readDiags.HasError(), "expected no error, got %v", readDiags)
	assert.Equal(t, "1:alice", d.Id())
}

func TestResourceMemberRead_MembershipAbsent(t *testing.T) {
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
						"users": []map[string]any{
							{"id": "bob"},
						},
					},
				},
			})
			return
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	r := resourceMember()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id": 1,
		"user_id":  "alice",
	})
	d.SetId("1:alice")

	readDiags := resourceMemberRead(context.Background(), d, client)
	require.False(t, readDiags.HasError(), "expected no error, got %v", readDiags)
	assert.Empty(t, d.Id(), "expected id to be cleared when user is not member of group")
}

func TestResourceMemberRead_EmptyUserId(t *testing.T) {
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
						"users": []map[string]any{
							{"id": "bob"},
						},
					},
				},
			})
			return
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	r := resourceMember()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"group_id": 1,
		"user_id":  "",
	})
	d.SetId("1:")

	readDiags := resourceMemberRead(context.Background(), d, client)
	require.False(t, readDiags.HasError(), "expected no error, got %v", readDiags)
	assert.Empty(t, d.Id(), "expected id to be cleared when user_id is empty and not in group")
}

func TestResourceMember_Description(t *testing.T) {
	r := resourceMember()
	assert.Equal(t, "Manages a LLDAP membership, i.e. a group-user relationship", r.Description)
}

