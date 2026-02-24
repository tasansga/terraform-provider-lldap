#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
tofu apply -auto-approve -var enable_out_of_band_user=true
tofu apply -auto-approve
cli_assert_state_change
tofu apply -auto-approve -destroy
