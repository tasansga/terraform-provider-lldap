#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var test_attribute_type="INTEGER" -var 'test_values=["42"]'
tofu apply -auto-approve -destroy -var test_attribute_type="INTEGER"

tofu apply -auto-approve -var test_attribute_type="DATE_TIME" -var 'test_values=["2023-01-01T00:00:00Z"]'
cli_assert_state_change
tofu apply -auto-approve -destroy -var test_attribute_type="DATE_TIME"
