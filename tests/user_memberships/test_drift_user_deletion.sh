#!/usr/bin/env bash

set -exo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

LLDAP_CLI_CMD="${LLDAP_CLI_BIN:-../../dist/lldap-cli}"

rm -f terraform.tfstate terraform.tfstate.backup *.tfplan

# Step 1: Initial apply to create user and memberships
tofu apply -auto-approve -var num_groups=5 -var max_groups=5
cli_assert_state_change

USER_ID=$(tofu output -json test_user | jq -r '.username')

# Step 2: Delete the underlying user entity out-of-band directly via lldap-cli
"$LLDAP_CLI_CMD" user delete "${USER_ID}"

# Step 3: Run tofu plan and verify it does not crash, and plans recreation
PLAN_OUTPUT=$(tofu plan -out=recreate.tfplan -json -var num_groups=5 -var max_groups=5)
PLAN_ACTIONS=$(echo "$PLAN_OUTPUT" | jq -r 'select(.type == "planned_change") | .change.action')

# Verify that recreation (create) is planned
CREATE_COUNT=$(echo "$PLAN_ACTIONS" | grep -c "create" || true)
[ "$CREATE_COUNT" -ge 1 ]

# Step 4: Apply the plan to recreate the deleted resources
tofu apply -auto-approve recreate.tfplan
cli_assert_state_change

# Step 5: Clean up
tofu apply -auto-approve -destroy -var num_groups=5 -var max_groups=5
rm -f recreate.tfplan
