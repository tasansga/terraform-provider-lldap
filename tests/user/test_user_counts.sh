#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

for count in 1 5; do
  tofu apply -auto-approve -var user_count=$count
cli_assert_state_change
  tofu apply -auto-approve -destroy -var user_count=$count
done
