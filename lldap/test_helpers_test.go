/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func newMockLldapServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *LldapClient) {
	server := httptest.NewServer(handler)
	parsedHttpUrl, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := &LldapClient{
		Config: Config{
			Context:  context.Background(),
			HttpUrl:  parsedHttpUrl,
			UserName: "admin",
			Password: "password",
			BaseDn:   "dc=example,dc=com",
		},
		Token:      "mock-token",
		HttpClient: server.Client(),
	}
	return server, client
}
