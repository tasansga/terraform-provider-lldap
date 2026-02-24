#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var create_user=true -var create_group=true
USER_ID=$(tofu output -json env_test_user | jq -r '.username')
GROUP_ID=$(tofu output -json env_test_group | jq -r '.id')

export LLDAP_HTTP_URL="http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}"
export LLDAP_LDAP_URL="ldap://${LLDAP_HOST}:${LLDAP_PORT_LDAP}"
export LLDAP_USER="admin"
export LLDAP_BASE_DN="dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com"
../../dist/lldap-cli user update "$USER_ID" --firstname "OutOfBand"

tofu plan
tofu apply -auto-approve
cli_assert_state_change
tofu plan

tofu apply -auto-approve -destroy -var create_user=true -var create_group=true
