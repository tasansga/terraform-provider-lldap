#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var test_is_list=true -var 'test_values=["list-value-1", "list-value-2", "list-value-3"]'
cli_assert_state_change
tofu apply -auto-approve -destroy -var test_is_list=true
