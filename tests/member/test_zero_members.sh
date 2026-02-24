#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var member_count=0
cli_assert_state_change
tofu apply -auto-approve -destroy -var member_count=0
