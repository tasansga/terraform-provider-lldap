#!/usr/bin/env bash

set -exo pipefail

echo "=== Group Attribute Assignment CRUD Lifecycle Test ==="

echo "=== Test Create ==="
tofu apply -auto-approve

echo "=== Test Read (via data source) ==="
tofu refresh

echo "=== Test Update (change values) ==="
tofu apply -auto-approve -var 'test_values=["updated-value-1"]'

echo "=== Test Delete ==="
tofu apply -auto-approve -destroy

echo "=== Test List Attribute Assignment ==="
tofu apply -auto-approve -var test_is_list=true -var 'test_values=["list-value-1", "list-value-2", "list-value-3"]'
tofu apply -auto-approve -destroy -var test_is_list=true

echo "=== Test Different Attribute Types ==="
# Test INTEGER attribute
tofu apply -auto-approve -var test_attribute_type="INTEGER" -var 'test_values=["42"]'
tofu apply -auto-approve -destroy -var test_attribute_type="INTEGER"

# Test DATE_TIME attribute
tofu apply -auto-approve -var test_attribute_type="DATE_TIME" -var 'test_values=["2023-01-01T00:00:00Z"]'
tofu apply -auto-approve -destroy -var test_attribute_type="DATE_TIME"

echo "=== Test Assignment Lifecycle ==="
# Create group and attribute
tofu apply -auto-approve
# Remove assignment only (keep group and attribute)
tofu apply -auto-approve -target=lldap_group_attribute_assignment.test_assignment -destroy
# Recreate assignment
tofu apply -auto-approve
# Clean up everything
tofu apply -auto-approve -destroy

echo "=== All group attribute assignment CRUD tests completed successfully! ==="
