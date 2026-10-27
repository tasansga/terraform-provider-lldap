#!/usr/bin/env bash

set -eo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

trap 'tofu apply -auto-approve -var create_ignored_password_user=true -destroy' EXIT

# Step 1: Initial apply to create resources
tofu apply -auto-approve -var create_ignored_password_user=true

USER_ID=$(tofu output -json ignored_password_user | jq -r '.username')
INITIAL_PASS=$(tofu output -raw ignored_user_password)
NEW_OOB_PASS="NewOutOfBandPassword123!"

login_status() {
  local user="$1"
  local pass="$2"
  jq -nc --arg u "$user" --arg p "$pass" '{"username": $u, "password": $p}' | \
    curl -s -o /dev/null -w "%{http_code}" -X POST "${LLDAP_HTTP_URL}/auth/simple/login" \
      -H "Content-Type: application/json" \
      --data-binary @-
}

# Step 2: Verify that initial password authenticates successfully against LLDAP
[ "$(login_status "$USER_ID" "$INITIAL_PASS")" -eq 200 ]

# Step 3: Change user password out-of-band using lldap-cli
"${LLDAP_CLI_BIN:-../../dist/lldap-cli}" user password "${USER_ID}" "${NEW_OOB_PASS}"

# Step 4: Verify out-of-band password works and initial password fails
[ "$(login_status "$USER_ID" "$NEW_OOB_PASS")" -eq 200 ]
[ "$(login_status "$USER_ID" "$INITIAL_PASS")" -ne 200 ]

# Step 5: Update a non-password attribute (email) via OpenTofu
tofu apply -auto-approve -var create_ignored_password_user=true -var ignored_password_email_prefix="updated"

# Step 6: Verify email was updated in LLDAP
CLI_OUTPUT=$("${LLDAP_CLI_BIN:-../../dist/lldap-cli}" user get "${USER_ID}")
ACTUAL_EMAIL=$(echo "$CLI_OUTPUT" | jq -r '.email')
echo "$ACTUAL_EMAIL" | grep -q "ign-updated-"

# Step 7: Verify out-of-band password was NOT clobbered by OpenTofu apply
[ "$(login_status "$USER_ID" "$NEW_OOB_PASS")" -eq 200 ]
[ "$(login_status "$USER_ID" "$INITIAL_PASS")" -ne 200 ]

cli_assert_state_change
