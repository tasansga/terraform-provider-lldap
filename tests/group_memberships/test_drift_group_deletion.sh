#!/usr/bin/env bash

set -exo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

LLDAP_CLI_CMD="${LLDAP_CLI_BIN:-../../dist/lldap-cli}"

rm -f terraform.tfstate terraform.tfstate.backup *.tfplan

# Step 1: Initial apply to create group and memberships
tofu apply -auto-approve -var num_users=5 -var max_users=5
cli_assert_state_change

GROUP_ID=$(tofu output -json test_group | jq -r '.id')

# Step 2: Delete the underlying group entity out-of-band directly via lldap-cli
"$LLDAP_CLI_CMD" group delete "${GROUP_ID}"

# Step 3: Run tofu plan and verify it does not crash, and plans recreation
PLAN_OUTPUT=$(tofu plan -out=recreate.tfplan -json -var num_users=5 -var max_users=5)
PLAN_ACTIONS=$(echo "$PLAN_OUTPUT" | jq -r 'select(.type == "planned_change") | .change.action')

# Verify that recreation (create) is planned
CREATE_COUNT=$(echo "$PLAN_ACTIONS" | grep -c "create" || true)
[ "$CREATE_COUNT" -ge 1 ]

# Step 4: Apply the plan to recreate the deleted resources
tofu apply -auto-approve recreate.tfplan
cli_assert_state_change

# Step 5: Clean up
tofu apply -auto-approve -destroy -var num_users=5 -var max_users=5
rm -f recreate.tfplan
