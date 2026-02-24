#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var create_group_with_members=false
tofu apply -auto-approve -var create_group_with_members=true
tofu apply -auto-approve -var create_test_users=false -var create_group_with_members=false
cli_assert_state_change
tofu apply -auto-approve -destroy -var create_test_users=false -var create_group_with_members=false
