#!/usr/bin/env bash
set -exo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"
rm -f terraform.tfstate terraform.tfstate.backup *.tfplan

LLDAP_CLI_CMD="${LLDAP_CLI_BIN:-../../dist/lldap-cli}"

# Create user using legacy `password` attribute
tofu apply -auto-approve -var password="hunter2"

USER_ID=$(tofu output -json username | jq -r '.')

# Expect no drift, plain text password is in state file
tofu show -json |
  jq -e '
    .values.root_module.resources[]
    | select(.address == "lldap_user.this")
    | .values.password == "hunter2"
  '

"$LLDAP_CLI_CMD" user verify-password "$USER_ID" "hunter2"
tofu plan -detailed-exitcode -var password="hunter2"

# Migrate to password_wo attribute
tofu apply -auto-approve -var password_wo="hunter3" -var password_wo_version="1"

# Expect no drift, no password in state file, new password works
tofu show -json |
  jq -e '
    .values.root_module.resources[]
    | select(.address == "lldap_user.this")
    | (.values.password == "")
  '

"$LLDAP_CLI_CMD" user verify-password "$USER_ID" "hunter3"

# Change password_wo input, without changing version, expect no change
tofu plan -detailed-exitcode -var password_wo="hunter4" -var password_wo_version="1"

# Change password_wo & password_wo_version, expect update, new password works
tofu apply -auto-approve -var password_wo="hunter4" -var password_wo_version="2"

tofu show -json |
  jq -e '
    .values.root_module.resources[]
    | select(.address == "lldap_user.this")
    | .values.password_wo_version == "2"
  '

"$LLDAP_CLI_CMD" user verify-password "$USER_ID" "hunter4"
tofu plan -detailed-exitcode -var password_wo="hunter4" -var password_wo_version="2"
