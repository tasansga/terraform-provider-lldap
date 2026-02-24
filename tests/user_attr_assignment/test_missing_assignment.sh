#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
USER_ID=$(tofu output -json test_user | jq -r '.username')
ATTR_ID=$(tofu output -json test_user_attr | jq -r '.name')

../../dist/lldap-cli attribute remove "${ATTR_ID}" "${USER_ID}" --user || true

ATTR_COUNT=$(../../dist/lldap-cli user get "${USER_ID}" | jq -r --arg attr "$ATTR_ID" '[.attributes[] | select(.name == $attr)] | length')
[ "$ATTR_COUNT" = "0" ]

PLAN_ACTIONS=$(tofu plan -out=plan2.tfplan -json 2>/dev/null | jq -r 'select(.type == "planned_change") | .change.action')
PLAN_OUTPUT=$(echo "$PLAN_ACTIONS" | grep -c "create" || echo "0")
[ "$PLAN_OUTPUT" -gt 0 ]

tofu apply -auto-approve plan2.tfplan
cli_assert_state_change

ATTR_COUNT=$(../../dist/lldap-cli user get "${USER_ID}" | jq -r --arg attr "$ATTR_ID" '[.attributes[] | select(.name == $attr)] | length')
[ "$ATTR_COUNT" -gt "0" ]

tofu apply -auto-approve -destroy
rm -f plan2.tfplan
