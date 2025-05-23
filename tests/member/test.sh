#!/usr/bin/env bash

set -exo pipefail

echo "=== Member CRUD Lifecycle Test ==="

echo "=== Test Create ==="
tofu apply -auto-approve

echo "=== Test Read (via data sources) ==="
tofu refresh

echo "=== Test Update (member count variations) ==="
for count in 1 5; do
  echo "=== Testing with $count members ==="
  tofu apply -auto-approve -var member_count=$count
  tofu apply -auto-approve -destroy -var member_count=$count
done

echo "=== Test Delete ==="
tofu apply -auto-approve
tofu apply -auto-approve -destroy

echo "=== Test Out-of-Band Member Removal ==="
tofu apply -auto-approve

echo "Removing member out of band using GraphQL API..."
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

echo "=== Test Drift Detection and Correction ==="
echo "Terraform should detect and recreate the removed member..."
tofu apply -auto-approve

echo "=== Clean up ==="
tofu apply -auto-approve -destroy

echo "=== Test Member Lifecycle with Different Configurations ==="
# Test with zero members
tofu apply -auto-approve -var member_count=0
tofu apply -auto-approve -destroy -var member_count=0

echo "=== All member CRUD tests completed successfully! ==="
