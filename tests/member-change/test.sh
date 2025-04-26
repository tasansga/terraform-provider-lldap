#!/usr/bin/env bash

set -exo pipefail

# Ensure plan does not change
tofu apply -auto-approve
tofu plan -detailed-exitcode

# Remove member out of band
auth_data="$(jo -- username="admin" password="$LLDAP_PASSWORD")"
auth_response="$(curl \
    --silent \
    --fail \
    --url "http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}/auth/simple/login" \
    --header 'Content-Type: application/json' \
    --data "$auth_data" \
)"
token="$(echo "$auth_response" | jq -r .token)"

query='{"operationName":"RemoveUserFromGroup","query":"mutation RemoveUserFromGroup($user: String!, $group: Int!) {removeUserFromGroup(userId: $user, groupId: $group) {ok}}"}'
variables="$(jo -- user="$(tofu output -raw user_id)" group="$(tofu output -raw group_id)")"
data="$(echo "$query" "{ \"variables\": $variables }" | jq -s add)"

curl \
    --silent \
    --fail \
    --url "http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}/api/graphql" \
    --header "Authorization: Bearer ${token}" \
    --header 'Content-Type: application/json' \
    --data "$data"

# Ensure member is recreated correctly
tofu apply -auto-approve
tofu plan -detailed-exitcode
tofu apply -auto-approve -destroy
