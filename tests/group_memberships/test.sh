#!/usr/bin/env bash

set -exo pipefail

echo "=== Group Memberships CRUD Lifecycle Test ==="

echo "=== Test Create (default 10 users) ==="
tofu apply -auto-approve

echo "=== Test Read (via data sources) ==="
tofu refresh

echo "=== Test Delete ==="
tofu apply -auto-approve -destroy

echo "=== Test with 0 users ==="
tofu apply -auto-approve -var num_users=0
tofu apply -auto-approve -destroy -var num_users=0

echo "=== Test Update (membership changes) ==="
tofu apply -auto-approve -var num_users=8
tofu apply -auto-approve -var num_users=5
tofu apply -auto-approve -destroy -var num_users=5

echo "=== Test Different User Counts ==="
for count in 1 3 7; do
  echo "=== Testing with $count users ==="
  tofu apply -auto-approve -var num_users=$count
  tofu apply -auto-approve -destroy -var num_users=$count
done

echo "=== Test Out-of-Band Changes ==="
tofu apply -auto-approve
tofu apply -auto-approve -var enable_out_of_band_user=true
tofu apply -auto-approve
tofu apply -auto-approve -destroy

echo "=== Test Member Count Variations ==="
# Test with different max_users settings
for max in 5 15; do
  echo "=== Testing with max_users=$max ==="
  tofu apply -auto-approve -var max_users=$max -var num_users=3
  tofu apply -auto-approve -destroy -var max_users=$max -var num_users=3
done

echo "=== All group memberships CRUD tests completed successfully! ==="
