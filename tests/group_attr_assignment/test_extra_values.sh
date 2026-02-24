#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
GROUP_ID=$(tofu output -json test_group | jq -r '.id')
ATTR_ID=$(tofu output -json test_group_attr | jq -r '.name')

../../dist/lldap-cli attribute add "${ATTR_ID}" "${GROUP_ID}" --group --values "extra-out-of-band-group-value"

tofu plan
tofu apply -auto-approve
cli_assert_state_change

tofu apply -auto-approve -destroy
