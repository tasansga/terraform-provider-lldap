#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"
rm -f terraform.tfstate terraform.tfstate.backup *.tfplan

tofu apply -no-color -auto-approve
tofu apply -no-color -auto-approve -var enable_out_of_band_user=true
tofu apply -no-color -auto-approve
cli_assert_state_change
tofu apply -no-color -auto-approve -destroy
