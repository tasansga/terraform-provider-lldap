#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

for max in 5 15; do
  tofu apply -auto-approve -var max_users=$max -var num_users=3
cli_assert_state_change
  tofu apply -auto-approve -destroy -var max_users=$max -var num_users=3
done
