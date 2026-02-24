#!/usr/bin/env bash

set -exo pipefail


source "$(dirname "${BASH_SOURCE[0]}")/../test_helpers/cli_checks.sh"

tofu apply -auto-approve -var test_attribute_type="INTEGER" -var 'test_values=["42"]'
tofu apply -auto-approve -destroy -var test_attribute_type="INTEGER"

tofu apply -auto-approve -var test_attribute_type="DATE_TIME" -var 'test_values=["2023-01-01T00:00:00Z"]'
tofu apply -auto-approve -destroy -var test_attribute_type="DATE_TIME"

tofu apply -auto-approve -var test_attribute_type="JPEG_PHOTO" -var 'test_values=["/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQH/2wBDAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQH/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwA/wA=="]'
cli_assert_state_change
tofu apply -auto-approve -destroy -var test_attribute_type="JPEG_PHOTO"
