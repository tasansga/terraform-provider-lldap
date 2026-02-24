#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var group_count=2 -var create_group_with_attrs=true -var create_test_users=true -var create_group_with_members=true
GROUP_IDS=$(tofu output -json created_groups | jq -r '.[].id')
FIRST_GROUP_ID=$(echo "$GROUP_IDS" | head -n1)
USER_ID=$(tofu output -json test_users | jq -r '.[0].username')

export LLDAP_BASE_DN="dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com"
export LLDAP_HTTP_URL="http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}"
export LLDAP_LDAP_URL="ldap://${LLDAP_HOST}:${LLDAP_PORT_LDAP}"
export LLDAP_USER="admin"

../../dist/lldap-cli group update "$FIRST_GROUP_ID" --displayname "Hacked Group Name"
tofu plan
tofu apply -auto-approve

../../dist/lldap-cli group delete "$FIRST_GROUP_ID"
tofu plan
tofu apply -auto-approve

if [ -n "$USER_ID" ]; then
    ../../dist/lldap-cli member remove "$FIRST_GROUP_ID" "$USER_ID" || echo "User not in group or command failed"
    tofu plan
    tofu apply -auto-approve
cli_assert_state_change
fi

tofu apply -auto-approve -destroy
