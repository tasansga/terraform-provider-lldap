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

# Ensure an user initially created without password can set one later
tofu apply -auto-approve
tofu apply -auto-approve -var nopasswd="yespasswd"
tofu plan -detailed-exitcode -var nopasswd="yespasswd"
tofu apply -auto-approve -destroy

# Ensure an user initially created with password can remove it later
tofu apply -auto-approve -var nopasswd="yespasswd"
tofu apply -auto-approve
tofu plan -detailed-exitcode
tofu apply -auto-approve -destroy
