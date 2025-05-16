#!/usr/bin/env bash

set -exo pipefail

echo "Ensure plan does not change and deletion works with default (10) users"
tofu apply -auto-approve
tofu plan -detailed-exitcode
tofu apply -auto-approve -destroy

echo "Ensure nothing happens with 0 users"
tofu apply -auto-approve -var num_users=0
# Looks like there's a bug (?) that displays empty terraform sets as `null`.
# Seems to be a known issue?
# - https://discuss.hashicorp.com/t/handling-of-nil-and-empty-schema-typelist/30594/2
#tofu plan -detailed-exitcode -var num_users=0
tofu apply -auto-approve -destroy

echo "Ensure updating memberships works"
tofu apply -auto-approve -var num_users=8
tofu apply -auto-approve -var num_users=5
tofu plan -detailed-exitcode -var num_users=5
tofu apply -auto-approve -destroy -var num_users=5

echo "Ensure out of band changes are handled"
tofu apply -auto-approve
tofu apply -auto-approve -var enable_out_of_band_user=true
tofu apply -auto-approve
tofu plan -detailed-exitcode
tofu apply -auto-approve -destroy
