#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

for i in {1..3}; do
  tofu apply -auto-approve
cli_assert_state_change
  tofu plan
  tofu apply -auto-approve -destroy
done
