#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var group_count=0
tofu apply -auto-approve -var group_count=3
tofu apply -auto-approve -var group_count=10
tofu apply -auto-approve -var group_count=5
tofu apply -auto-approve -var group_count=1
cli_assert_state_change
tofu apply -auto-approve -destroy -var group_count=1
