#!/usr/bin/env bash

set -exo pipefail

echo "=== User Memberships CRUD Lifecycle Test ==="

echo "=== Test Create (default 10 groups) ==="
tofu apply -auto-approve

echo "=== Test Read (via data sources) ==="
tofu refresh

echo "=== Test Delete ==="
tofu apply -auto-approve -destroy

echo "=== Test with 0 groups ==="
tofu apply -auto-approve -var num_groups=0
tofu apply -auto-approve -destroy -var num_groups=0

echo "=== Test Update (membership changes) ==="
tofu apply -auto-approve -var num_groups=8
tofu apply -auto-approve -var num_groups=5
tofu apply -auto-approve -destroy -var num_groups=5

echo "=== Test Different Group Counts ==="
for count in 1 3 7; do
  echo "=== Testing with $count groups ==="
  tofu apply -auto-approve -var num_groups=$count
  tofu apply -auto-approve -destroy -var num_groups=$count
done

echo "=== Test Group Count Variations ==="
# Test with different max_groups settings
for max in 5 15; do
  echo "=== Testing with max_groups=$max ==="
  tofu apply -auto-approve -var max_groups=$max -var num_groups=3
  tofu apply -auto-approve -destroy -var max_groups=$max -var num_groups=3
done

echo "=== All user memberships CRUD tests completed successfully! ==="
