#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var create_group_attr=false
tofu apply -auto-approve -var create_group_attr=true
tofu apply -auto-approve -var create_group_attr=false
cli_assert_state_change
tofu apply -auto-approve -destroy -var create_group_attr=false
