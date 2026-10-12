#!/usr/bin/env bash

set -exo pipefail


tofu apply -auto-approve

# Import the member again, and make sure that the read sets all the attributes,
# so the next plan has no changes.
member_id="$(tofu output -json member | jq -r .id)"
tofu state rm lldap_member.member
tofu import lldap_member.member "$member_id"
tofu plan -detailed-exitcode

tofu apply -auto-approve -destroy
