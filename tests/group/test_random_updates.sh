#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
tofu taint random_string.suffix
tofu apply -auto-approve
cli_assert_state_change
tofu apply -auto-approve -destroy
