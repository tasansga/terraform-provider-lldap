#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

for count in 1 3 7; do
  tofu apply -auto-approve -var num_groups=$count
cli_assert_state_change
  tofu apply -auto-approve -destroy -var num_groups=$count
done
