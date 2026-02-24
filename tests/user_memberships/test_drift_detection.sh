#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

rm -f terraform.tfstate terraform.tfstate.backup *.tfplan
tofu apply -auto-approve -var num_groups=5 -var max_groups=10
USER_ID=$(tofu output -json test_user | jq -r '.username')
ALL_GROUPS=$(tofu output -json all_groups | jq -r '.[].id')
MANAGED_GROUPS=$(tofu output -json user_memberships | jq -r '.group_ids[]')

FIRST_MANAGED_GROUP=$(echo "${MANAGED_GROUPS}" | head -n1)
../../dist/lldap-cli member remove "${FIRST_MANAGED_GROUP}" "${USER_ID}"

PLAN_ACTIONS=$(tofu plan -out=plan2.tfplan -json -var num_groups=5 -var max_groups=10 2>/dev/null | jq -r 'select(.type == "planned_change") | .change.action')
if echo "$PLAN_ACTIONS" | grep -q "update"; then
    PLAN_OUTPUT=$(echo "$PLAN_ACTIONS" | grep -c "update")
else
    PLAN_OUTPUT=0
fi
[ "$PLAN_OUTPUT" -gt 0 ]

tofu apply -auto-approve plan2.tfplan
cli_assert_state_change

USER_GROUPS_AFTER=$(../../dist/lldap-cli user get "${USER_ID}" | jq -r '.groups[]?.id' | sort)
EXPECTED_GROUPS=$(echo -e "${MANAGED_GROUPS}" | sort)
[ "$USER_GROUPS_AFTER" = "$EXPECTED_GROUPS" ]

tofu apply -auto-approve -destroy
rm -f plan.tfplan plan2.tfplan
