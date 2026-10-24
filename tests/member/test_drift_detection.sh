#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
auth_data="$(jq -n --arg username "admin" --arg password "$LLDAP_PASSWORD" '{username: $username, password: $password}')"
auth_response="$(curl \
    --silent \
    --fail \
    --url "http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}/auth/simple/login" \
    --header 'Content-Type: application/json' \
    --data "$auth_data" \
)"
token="$(echo "$auth_response" | jq -r .token)"

query='{"operationName":"RemoveUserFromGroup","query":"mutation RemoveUserFromGroup($user: String!, $group: Int!) {removeUserFromGroup(userId: $user, groupId: $group) {ok}}"}'
variables="$(jq -n --arg user "$(tofu output -raw user_id)" --arg group "$(tofu output -raw group_id)" '{user: $user, group: ($group | tonumber)}')"
data="$(echo "$query" "{ \"variables\": $variables }" | jq -s add)"

curl \
    --silent \
    --fail \
    --url "http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}/api/graphql" \
    --header "Authorization: Bearer ${token}" \
    --header 'Content-Type: application/json' \
    --data "$data"

tofu apply -auto-approve
cli_assert_state_change
tofu apply -auto-approve -destroy
