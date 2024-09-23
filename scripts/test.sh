#!/usr/bin/env bash

set -eo pipefail

if [[ -z ${MAKE_TERMOUT+x} ]]
then
    echo 'Do not call this script directly.'
    echo 'Instead, use: `make test`'
    exit 1
fi

function on_exit {
    docker stop "$cnt_id" || true
    if [[ -d "$temp_test_dir" ]]
    then
        rm -vRf "$temp_test_dir"
    fi
}

trap on_exit EXIT

scripts_dir=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)
tf_provider_lldap_root_dir=$(realpath "${scripts_dir}/..")
temp_test_dir=$(mktemp -d)

cnt_id=$(docker run \
    --detach \
    --rm \
    --env LLDAP_LDAP_USER_PASS=this_is_a_very_safe_password \
    lldap/lldap:stable)

cnt_ip=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$cnt_id")
sleep 3 # Need to wait for the container to be ready

LLDAP_HOST="$cnt_ip" LLDAP_PASSWORD="this_is_a_very_safe_password" go test ./lldap

cd "${tf_provider_lldap_root_dir}"
mkdir -p "${temp_test_dir}/plugins/registry.opentofu.org/tasansga/lldap/0.0.1/linux_amd64/"
cp "${tf_provider_lldap_root_dir}/dist/tf-provider-lldap" "${temp_test_dir}/plugins/registry.opentofu.org/tasansga/lldap/0.0.1/linux_amd64/terraform-provider-lldap"

for f in $tf_provider_lldap_root_dir/tests/*
do
    if [[ -d "$f" ]]
    then
        cd "$f"
        rm -Rvf .terraform .terraform.lock.hcl
        tofu init -reconfigure -upgrade -plugin-dir="${temp_test_dir}/plugins"
        tofu test -var "lldap_url=http://${cnt_ip}:17170" -var 'lldap_username=admin' -var 'lldap_password=this_is_a_very_safe_password'
    fi
done
