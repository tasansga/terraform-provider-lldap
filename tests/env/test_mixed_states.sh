#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var create_user=true -var create_group=false
tofu apply -auto-approve -var create_user=true -var create_group=true
tofu plan -var create_user=true -var create_group=true
tofu apply -auto-approve -var create_user=false -var create_group=true
cli_assert_state_change
tofu plan -var create_user=false -var create_group=true
tofu apply -auto-approve -destroy -var create_user=false -var create_group=true
