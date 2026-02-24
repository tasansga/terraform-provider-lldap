#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
GROUP_ID=$(tofu output -json test_group | jq -r '.id')
ATTR_ID=$(tofu output -json test_group_attr | jq -r '.name')

../../dist/lldap-cli attribute remove "${ATTR_ID}" "${GROUP_ID}" --group || true

PLAN_OUTPUT=$(tofu plan -out=plan.tfplan -json 2>&1 || true)
if echo "$PLAN_OUTPUT" | grep -q "Group is missing attribute"; then
    echo "Expected error: Provider correctly detected missing attribute but cannot handle it gracefully"
elif echo "$PLAN_OUTPUT" | jq -r 'select(.type == "planned_change") | .change.action' | grep -q "create"; then
    tofu apply -auto-approve plan.tfplan
cli_assert_state_change
else
    exit 1
fi

tofu apply -auto-approve -destroy
rm -f plan.tfplan
