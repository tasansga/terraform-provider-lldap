#!/usr/bin/env bash

set -exo pipefail

# Ensure plan does not change and deletion works
tofu apply -auto-approve
tofu plan -detailed-exitcode
tofu apply -auto-approve -destroy

# Ensure password can be updated
tofu apply -auto-approve
tofu taint random_password.user
tofu apply -auto-approve
tofu apply -auto-approve -destroy

# Ensure email can be updated
tofu apply -auto-approve
tofu taint random_string.email_prefix
tofu apply -auto-approve
tofu apply -auto-approve -destroy
