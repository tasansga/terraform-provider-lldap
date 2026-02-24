#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
USER_ID=$(tofu output -json main_user | jq -r '.username')

../../dist/lldap-cli user update "${USER_ID}" \
    --displayname "Out-of-Band Display Name" \
    --firstname "OutOfBand" \
    --lastname "User"

CLI_OUTPUT=$(../../dist/lldap-cli user get "${USER_ID}")
ACTUAL_DISPLAY_NAME=$(echo "$CLI_OUTPUT" | jq -r '.displayName')
[ "$ACTUAL_DISPLAY_NAME" = "Out-of-Band Display Name" ]

PLAN_OUTPUT=$(tofu plan -out=plan.tfplan -json 2>/dev/null | jq -r 'select(.type == "planned_change") | .change.action' | grep -c "update" || echo "0")
[ "$PLAN_OUTPUT" -gt 0 ]

tofu apply -auto-approve plan.tfplan

EXPECTED_DISPLAY_NAME=$(tofu output -json main_user | jq -r '.display_name')
ACTUAL_DISPLAY_NAME=$(../../dist/lldap-cli user get "${USER_ID}" | jq -r '.displayName')
[ "$ACTUAL_DISPLAY_NAME" = "$EXPECTED_DISPLAY_NAME" ]

../../dist/lldap-cli user password "${USER_ID}" "OutOfBandPassword123!"
tofu plan
tofu apply -auto-approve

../../dist/lldap-cli user delete "${USER_ID}"

! ../../dist/lldap-cli user get "${USER_ID}" 2>/dev/null

PLAN_OUTPUT=$(tofu plan -out=plan3.tfplan -json 2>/dev/null | jq -r 'select(.type == "planned_change") | .change.action' | grep -c "create" || echo "0")
[ "$PLAN_OUTPUT" -gt 0 ]

tofu apply -auto-approve plan3.tfplan
cli_assert_state_change

../../dist/lldap-cli user get "${USER_ID}" >/dev/null 2>&1

tofu apply -auto-approve -destroy
rm -f plan.tfplan plan2.tfplan plan3.tfplan
