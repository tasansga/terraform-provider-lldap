/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

func TestGetUserAttributeSchema_CaseInsensitive(t *testing.T) {
	server, client := newMockLldapServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/graphql" {
			var query struct {
				OperationName string `json:"operationName"`
			}
			err := json.NewDecoder(r.Body).Decode(&query)
			require.NoError(t, err)

			if query.OperationName == "GetUserAttributesSchema" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"schema": map[string]any{
							"userSchema": map[string]any{
								"attributes": []map[string]any{
									{
										"name":          "sshpublickey",
										"attributeType": "STRING",
										"isList":        true,
										"isVisible":     true,
										"isEditable":    false,
										"isHardcoded":   false,
										"isReadonly":    false,
									},
								},
							},
						},
					},
				})
				return
			}
		}

		http.Error(w, "not found", http.StatusNotFound)
	})
	defer server.Close()

	schema, diags := client.GetUserAttributeSchema("sshPublicKey")
	require.False(t, diags.HasError())
	require.NotNil(t, schema, "expected case-insensitive match for sshPublicKey")
	assert.Equal(t, "sshpublickey", schema.Name)
}

func TestGetGroupAttributeSchema_CaseInsensitive(t *testing.T) {
	server, client := newMockLldapServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/graphql" {
			var query struct {
				OperationName string `json:"operationName"`
			}
			err := json.NewDecoder(r.Body).Decode(&query)
			require.NoError(t, err)

			if query.OperationName == "GetGroupAttributesSchema" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"schema": map[string]any{
							"groupSchema": map[string]any{
								"attributes": []map[string]any{
									{
										"name":          "customgroupattr",
										"attributeType": "STRING",
										"isList":        false,
										"isVisible":     true,
										"isHardcoded":   false,
										"isReadonly":    false,
									},
								},
							},
						},
					},
				})
				return
			}
		}

		http.Error(w, "not found", http.StatusNotFound)
	})
	defer server.Close()

	schema, diags := client.GetGroupAttributeSchema("CustomGroupAttr")
	require.False(t, diags.HasError())
	require.NotNil(t, schema, "expected case-insensitive match for CustomGroupAttr")
	assert.Equal(t, "customgroupattr", schema.Name)
}

func TestSetUserPassword_Concurrency(t *testing.T) {
	origFunc := getLdapBindConnectionFunc
	defer func() { getLdapBindConnectionFunc = origFunc }()

	c1, c2 := net.Pipe()
	_ = c2.Close()
	defer func() { _ = c1.Close() }()
	ldapConn := ldap.NewConn(c1, false)
	ldapConn.Start()

	getLdapBindConnectionFunc = func(ldapUrl string, baseDn string, username string, password string) (*ldap.Conn, diag.Diagnostics) {
		time.Sleep(10 * time.Millisecond)
		return ldapConn, nil
	}

	client := &LldapClient{
		Config: Config{
			LdapUrl:  &url.URL{Scheme: "ldap", Host: "127.0.0.1:389"},
			UserName: "admin",
			Password: "password",
			BaseDn:   "dc=example,dc=com",
		},
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = client.SetUserPassword(fmt.Sprintf("user%d", idx), "newpass")
		}(i)
	}
	wg.Wait()
}

func TestEscapeDN_SpecialCharacters(t *testing.T) {
	escaped := ldap.EscapeDN("user,test+special<val>")
	assert.Equal(t, `user\,test\+special\<val\>`, escaped)
}


