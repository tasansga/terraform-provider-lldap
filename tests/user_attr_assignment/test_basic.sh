#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
tofu refresh
tofu apply -auto-approve -var 'test_values=["updated-value-1"]'
cli_assert_state_change
tofu apply -auto-approve -destroy
