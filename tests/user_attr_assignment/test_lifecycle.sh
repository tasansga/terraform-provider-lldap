#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve
tofu apply -auto-approve -target=lldap_user_attribute_assignment.test_assignment -destroy
tofu apply -auto-approve
cli_assert_state_change
tofu apply -auto-approve -destroy
