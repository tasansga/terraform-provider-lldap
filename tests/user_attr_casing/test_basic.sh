#!/usr/bin/env bash

set -exo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

# Test mixed-case / camelCase attribute name creation, refresh, and plan idempotency
tofu apply -auto-approve
tofu plan -detailed-exitcode
tofu refresh
tofu plan -detailed-exitcode
cli_assert_state_change
tofu apply -auto-approve -destroy
