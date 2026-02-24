#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var test_attribute_type="INTEGER"
tofu apply -auto-approve -destroy -var test_attribute_type="INTEGER"

tofu apply -auto-approve -var test_attribute_type="DATE_TIME"
tofu apply -auto-approve -destroy -var test_attribute_type="DATE_TIME"

tofu apply -auto-approve -var test_attribute_type="JPEG_PHOTO"
cli_assert_state_change
tofu apply -auto-approve -destroy -var test_attribute_type="JPEG_PHOTO"
