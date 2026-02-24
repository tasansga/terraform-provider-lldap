#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var group_count=3 -var create_group_with_attrs=true -var create_test_users=true -var create_group_with_members=true
tofu apply -auto-approve -var group_count=3 -var create_group_with_attrs=true -var create_test_users=true -var create_group_with_members=false
tofu apply -auto-approve -var group_count=3 -var create_group_with_attrs=false -var create_test_users=true -var create_group_with_members=false
tofu apply -auto-approve -var group_count=1 -var create_group_with_attrs=false -var create_test_users=true -var create_group_with_members=false
tofu apply -auto-approve -var group_count=1 -var create_group_with_attrs=false -var create_test_users=false -var create_group_with_members=false
cli_assert_state_change
tofu apply -auto-approve -destroy
