#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"
LLDAP_CLI_CMD="${LLDAP_CLI_BIN:-../../dist/lldap-cli}"
rm -f terraform.tfstate terraform.tfstate.backup *.tfplan

tofu apply -no-color -auto-approve

GROUP_ID=$(tofu output -json test_group | jq -r '.id')
OOB_USER_ID=$(tofu output -json out_of_band_user | jq -r '.id')
"$LLDAP_CLI_CMD" member add "${GROUP_ID}" "${OOB_USER_ID}"

tofu apply -no-color -auto-approve
cli_assert_state_change
tofu apply -no-color -auto-approve -destroy
