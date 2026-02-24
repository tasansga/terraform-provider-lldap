#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
USER_ID=$(tofu output -json test_user | jq -r '.username')
GROUP_ID=$(tofu output -json test_group | jq -r '.id')

export LLDAP_BASE_DN="dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com"
export LLDAP_HTTP_URL="http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}"
export LLDAP_LDAP_URL="ldap://${LLDAP_HOST}:${LLDAP_PORT_LDAP}"
export LLDAP_USER="admin"

../../dist/lldap-cli user update "$USER_ID" --email "changed@example.com"
tofu plan
tofu apply -auto-approve

../../dist/lldap-cli group update "$GROUP_ID" --displayname "Changed Group Name"
tofu plan
tofu apply -auto-approve

../../dist/lldap-cli user delete "$USER_ID"
tofu plan
tofu apply -auto-approve
cli_assert_state_change

tofu apply -auto-approve -destroy
