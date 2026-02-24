#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
USER_ID=$(tofu output -json test_user | jq -r '.username')
ATTR_ID=$(tofu output -json test_user_attr | jq -r '.name')

../../dist/lldap-cli attribute remove "${ATTR_ID}" "${USER_ID}" --user || true
../../dist/lldap-cli attribute add "${ATTR_ID}" "${USER_ID}" --user --values "out-of-band-value"

ACTUAL_VALUE=$(../../dist/lldap-cli user get "${USER_ID}" | jq -r --arg attr "$ATTR_ID" '.attributes[] | select(.name == $attr) | .value[0]')
[ "$ACTUAL_VALUE" = "out-of-band-value" ]

PLAN_ACTIONS=$(tofu plan -out=plan.tfplan -json 2>/dev/null | jq -r 'select(.type == "planned_change") | .change.action')
PLAN_OUTPUT=$(echo "$PLAN_ACTIONS" | grep -c "update" || echo "0")
[ "$PLAN_OUTPUT" -gt 0 ]

tofu apply -auto-approve plan.tfplan
cli_assert_state_change

EXPECTED_VALUE=$(tofu output -json test_assignment | jq -r '.value[0]')
ACTUAL_VALUE=$(../../dist/lldap-cli user get "${USER_ID}" | jq -r --arg attr "$ATTR_ID" '.attributes[] | select(.name == $attr) | .value[0]')
[ "$ACTUAL_VALUE" = "$EXPECTED_VALUE" ]

tofu apply -auto-approve -destroy
rm -f plan.tfplan
