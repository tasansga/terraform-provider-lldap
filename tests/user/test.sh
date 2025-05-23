#!/usr/bin/env bash

set -exo pipefail

echo "=== User CRUD Lifecycle Test ==="

echo "=== Test Create ==="
tofu apply -auto-approve

echo "=== Test Read (via data sources) ==="
tofu refresh

echo "=== Test Update (password change) ==="
tofu taint random_password.user
tofu apply -auto-approve

echo "=== Test Update (email change) ==="
tofu taint random_string.email_prefix
tofu apply -auto-approve

echo "=== Test Delete ==="
tofu apply -auto-approve -destroy

echo "=== Test User Without Password ==="
tofu apply -auto-approve

echo "=== Test Adding Password to User ==="
tofu apply -auto-approve -var nopasswd="yespasswd"

echo "=== Test Removing Password from User ==="
tofu apply -auto-approve

echo "=== Clean up ==="
tofu apply -auto-approve -destroy

echo "=== Test User Count Variations ==="
# Test with different user counts
for count in 1 5; do
  echo "=== Testing with $count users ==="
  tofu apply -auto-approve -var user_count=$count
  tofu apply -auto-approve -destroy -var user_count=$count
done

echo "=== Test User with Custom Attributes ==="
tofu apply -auto-approve -var create_user_with_attrs=true
tofu apply -auto-approve -destroy -var create_user_with_attrs=true

echo "=== Test User without Custom Attributes ==="
tofu apply -auto-approve -var create_user_with_attrs=false
tofu apply -auto-approve -destroy -var create_user_with_attrs=false

echo "=== All user CRUD tests completed successfully! ==="
