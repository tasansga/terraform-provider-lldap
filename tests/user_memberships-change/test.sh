#!/usr/bin/env bash

set -exo pipefail

echo "Ensure plan does not change and deletion works with default (10) groups"
tofu apply -auto-approve
tofu plan -detailed-exitcode
tofu apply -auto-approve -destroy

echo "Ensure nothing happens with 0 groups"
tofu apply -auto-approve -var num_groups=0
# Looks like there's a bug (?) that displays empty terraform sets as `null`.
# Seems to be a known issue?
# - https://discuss.hashicorp.com/t/handling-of-nil-and-empty-schema-typelist/30594/2
#tofu plan -detailed-exitcode -var num_groups=0
tofu apply -auto-approve -destroy -var num_groups=0

echo "Ensure updating memberships works"
tofu apply -auto-approve -var num_groups=8
tofu apply -auto-approve -var num_groups=5
tofu plan -detailed-exitcode -var num_groups=5
tofu apply -auto-approve -destroy -var num_groups=5
