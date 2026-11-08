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

	"github.com/hashicorp/go-cty/cty"
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

func TestResourceUserUpdate_PasswordWoVersionUnset(t *testing.T) {
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
					"updateUser": map[string]any{
						"ok": true,
					},
				},
			})
			return
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	r := resourceUser()
	s := &terraform.InstanceState{
		ID: "testuser",
		Attributes: map[string]string{
			"id":                  "testuser",
			"username":            "testuser",
			"email":               "user@example.com",
			"password_wo_version": "1",
		},
	}
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"password_wo_version": {
				Old: "1",
				New: "",
			},
		},
	}

	_, diags := r.Apply(context.Background(), s, diff, client)
	require.False(t, diags.HasError(), "unsetting password_wo_version should not return error: %v", diags)
}

func TestResourceUserCreate_SetIdCalledWhenSetPasswordFails(t *testing.T) {
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

			if query.OperationName == "CreateUser" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"createUser": map[string]any{
							"id":           "newuser",
							"creationDate": "2026-01-01",
							"uuid":         "mock-uuid",
						},
					},
				})
				return
			}

			if query.OperationName == "GetUserDetails" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"user": map[string]any{
							"id":           "newuser",
							"email":        "newuser@example.com",
							"creationDate": "2026-01-01",
							"uuid":         "mock-uuid",
						},
					},
				})
				return
			}
		}

		http.Error(w, fmt.Sprintf("unhandled request: %s %s", r.Method, r.URL.Path), http.StatusNotFound)
	})
	defer server.Close()

	dummyLdapUrl, err := url.Parse("ldap://127.0.0.1:1")
	require.NoError(t, err)
	client.Config.LdapUrl = dummyLdapUrl

	r := resourceUser()
	s := &terraform.InstanceState{}
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"username": {
				New: "newuser",
			},
			"email": {
				New: "newuser@example.com",
			},
			"password": {
				New: "newpass",
			},
		},
	}

	newState, diags := r.Apply(context.Background(), s, diff, client)
	require.True(t, diags.HasError(), "expected error from SetUserPassword")
	require.NotNil(t, newState, "expected partial state to be returned")
	assert.Equal(t, "newuser", newState.ID, "expected ID to be set even if SetUserPassword fails")
	assert.Equal(t, "newuser@example.com", newState.Attributes["email"], "expected email to be saved in partial state")
}

func TestResourceUserUpdate_PasswordWoVersionChanged_EmptyPasswordFails(t *testing.T) {
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

	r := resourceUser()
	s := &terraform.InstanceState{
		ID: "testuser",
		Attributes: map[string]string{
			"id":                  "testuser",
			"username":            "testuser",
			"email":               "user@example.com",
			"password_wo_version": "1",
		},
	}
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"password_wo_version": {
				Old: "1",
				New: "2",
			},
		},
	}

	_, diags := r.Apply(context.Background(), s, diff, client)
	require.True(t, diags.HasError(), "expected error when password_wo_version changes but password_wo is empty")
	assert.Contains(t, diags[0].Summary, "password_wo must be non-empty when password_wo_version changes")
}

func TestResourceUserGetWriteOnlyPassword_MarkedSensitiveValue(t *testing.T) {
	r := resourceUser()
	s := &terraform.InstanceState{
		ID: "testuser",
		Attributes: map[string]string{
			"id":       "testuser",
			"username": "testuser",
			"email":    "user@example.com",
		},
	}
	markedVal := cty.StringVal("mysecretpassword").Mark("sensitive")
	diff := &terraform.InstanceDiff{
		RawConfig: cty.ObjectVal(map[string]cty.Value{
			"password_wo":         markedVal,
			"password_wo_version": cty.StringVal("1"),
		}),
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"password_wo_version": {
				Old: "",
				New: "1",
			},
		},
	}

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
					"updateUser": map[string]any{
						"ok": true,
					},
				},
			})
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	})
	defer server.Close()

	dummyLdapUrl, err := url.Parse("ldap://127.0.0.1:1")
	require.NoError(t, err)
	client.Config.LdapUrl = dummyLdapUrl

	// If marked value is not unmarked, AsString() unconditionally panics!
	_, diags := r.Apply(context.Background(), s, diff, client)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "127.0.0.1:1")
}




