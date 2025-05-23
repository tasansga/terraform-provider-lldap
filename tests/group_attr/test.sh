#!/usr/bin/env bash

set -exo pipefail

echo "=== Group Attribute CRUD Lifecycle Test ==="

echo "=== Test Create ==="
tofu apply -auto-approve

echo "=== Test Read (via data source) ==="
tofu refresh

echo "=== Test Update (taint and recreate) ==="
tofu taint lldap_group_attribute.test_attr
tofu apply -auto-approve

echo "=== Test Delete ==="
tofu apply -auto-approve -destroy

echo "=== Test Multiple Attribute Types ==="
# Test different attribute types (valid for group attributes: DATE_TIME, INTEGER, JPEG_PHOTO, STRING)
tofu apply -auto-approve -var test_attribute_type="INTEGER"
tofu apply -auto-approve -destroy -var test_attribute_type="INTEGER"

tofu apply -auto-approve -var test_attribute_type="DATE_TIME"
tofu apply -auto-approve -destroy -var test_attribute_type="DATE_TIME"

tofu apply -auto-approve -var test_attribute_type="JPEG_PHOTO"
tofu apply -auto-approve -destroy -var test_attribute_type="JPEG_PHOTO"

echo "=== Test List Attributes ==="
tofu apply -auto-approve -var test_is_list=true
tofu apply -auto-approve -destroy -var test_is_list=true

echo "=== Test Visibility Settings ==="
tofu apply -auto-approve -var test_is_visible=false
tofu apply -auto-approve -destroy -var test_is_visible=false

echo "=== All group attribute CRUD tests completed successfully! ==="
