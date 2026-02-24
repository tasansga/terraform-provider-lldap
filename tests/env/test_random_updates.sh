#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
tofu taint random_string.suffix
tofu apply -auto-approve
tofu plan
tofu taint random_password.password
tofu apply -auto-approve
cli_assert_state_change
tofu plan
tofu apply -auto-approve -destroy
