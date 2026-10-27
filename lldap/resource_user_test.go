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
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceUserUpdate_PasswordNotChanged(t *testing.T) {
	var updateUserCalled int32

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

			if query.OperationName == "UpdateUser" {
				atomic.AddInt32(&updateUserCalled, 1)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"updateUser": map[string]any{
							"ok": true,
						},
					},
				})
				return
			}
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	// If SetUserPassword or IsValidPassword is called via LDAP, since client.Config.LdapUrl is nil / unconfigured,
	// or if it attempts LDAP bind, it would fail or panic if called!
	// Specifically, in newMockLldapServer, LdapUrl is not set (nil).
	// Therefore, any call to lc.IsValidPassword or lc.SetUserPassword will fail/panic trying to dial LDAP.

	r := resourceUser()
	s := &terraform.InstanceState{
		ID: "testuser",
		Attributes: map[string]string{
			"id":       "testuser",
			"username": "testuser",
			"email":    "old@example.com",
			"password": "initialpassword",
		},
	}
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"email": {
				Old: "old@example.com",
				New: "new@example.com",
			},
		},
	}

	newState, diags := r.Apply(context.Background(), s, diff, client)
	require.False(t, diags.HasError(), "expected apply without error when password not changed: %v", diags)
	assert.Equal(t, int32(1), atomic.LoadInt32(&updateUserCalled))
	assert.Equal(t, "new@example.com", newState.Attributes["email"])
}

func TestResourceUserUpdate_PasswordChanged_CallsSetUserPassword(t *testing.T) {
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

			if query.OperationName == "UpdateUser" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"updateUser": map[string]any{
							"ok": true,
						},
					},
				})
				return
			}
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	// Configure a dummy LDAP URL so SetUserPassword attempts connection and returns a diagnostic error rather than panicking
	dummyLdapUrl, err := url.Parse("ldap://127.0.0.1:1")
	require.NoError(t, err)
	client.Config.LdapUrl = dummyLdapUrl

	r := resourceUser()
	s := &terraform.InstanceState{
		ID: "testuser",
		Attributes: map[string]string{
			"id":       "testuser",
			"username": "testuser",
			"email":    "user@example.com",
			"password": "oldpassword",
		},
	}
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"password": {
				Old: "oldpassword",
				New: "newpassword",
			},
		},
	}

	// Because password changed in diff, resourceUserUpdate invokes SetUserPassword,
	// which attempts LDAP connection to the dummy URL and returns an LDAP dial diagnostic error.
	_, diags := r.Apply(context.Background(), s, diff, client)
	require.True(t, diags.HasError(), "expected error from SetUserPassword when dialing dummy LDAP URL")
	assert.Contains(t, diags[0].Summary, "127.0.0.1:1")
}
