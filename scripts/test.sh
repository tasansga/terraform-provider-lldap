#!/usr/bin/env bash

set -eo pipefail

function start_lldap_server {
    local passwd=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 18; echo)
    cnt_id=$(docker run \
        --detach \
        --rm \
        --env "LLDAP_LDAP_USER_PASS=${passwd}" \
        lldap/lldap:stable)
    local cnt_ip=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$cnt_id")

    sleep 3 # Need to wait for the container to be ready

    cat <<EOF
export LLDAP_CONTAINER_ID="$cnt_id"
export LLDAP_HOST="$cnt_ip"
export LLDAP_PASSWORD="$passwd"
EOF
    export LLDAP_CONTAINER_ID="$cnt_id"
    export LLDAP_HOST="$cnt_ip"
    export LLDAP_PASSWORD="$passwd"
}

function stop_lldap_server {
    docker stop "$LLDAP_CONTAINER_ID" || true
    cat <<EOF
unset LLDAP_CONTAINER_ID
unset LLDAP_HOST
unset LLDAP_PASSWORD
EOF
    unset LLDAP_CONTAINER_ID
    unset LLDAP_HOST
    unset LLDAP_PASSWORD
}

function run_unit_test {
    start_lldap_server
    trap stop_lldap_server RETURN
    trap stop_lldap_server EXIT

    go test ./lldap
}

function run_integration_test {
    local test_path="$1"
    start_lldap_server
    trap stop_lldap_server RETURN
    trap stop_lldap_server EXIT

    echo "Running test: ${test_path}"
    cd "$test_path"
    rm -Rvf .terraform .terraform.lock.hcl terraform.tfstate terraform.tfstate.backup
    cat > "$test_path/.tfvars" << EOF
lldap_url="http://${LLDAP_HOST}:17170"
lldap_username="admin"
lldap_password="$LLDAP_PASSWORD"
EOF
    tofu init -reconfigure -upgrade
    tofu test -var-file="$test_path/.tfvars"
}

function run_integration_tests {
    local scripts_dir=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)
    local tf_provider_lldap_root_dir=$(realpath "${scripts_dir}/..")
    temp_test_dir=$(mktemp -d)

    function on_integration_test_exit {
        if [[ -d "$temp_test_dir" ]]
        then
            rm -vRf "$temp_test_dir"
        fi
    }
    trap on_integration_test_exit RETURN
    trap on_integration_test_exit EXIT

    cd "${tf_provider_lldap_root_dir}"
    mkdir -p "${temp_test_dir}/plugins/registry.opentofu.org/tasansga/lldap/0.0.1/linux_amd64/"
    cp "${tf_provider_lldap_root_dir}/dist/tf-provider-lldap" "${temp_test_dir}/plugins/registry.opentofu.org/tasansga/lldap/0.0.1/linux_amd64/terraform-provider-lldap"

    export TF_CLI_CONFIG_FILE="${temp_test_dir}/test.tfrc"
    cat > "$TF_CLI_CONFIG_FILE" << EOF
provider_installation {
  filesystem_mirror {
    path    = "${temp_test_dir}/plugins"
    include = ["tasansga/lldap"]
  }
  direct {
    exclude = ["tasansga/lldap"]
  }
}
EOF

    for f in $tf_provider_lldap_root_dir/tests/*
    do
        if [[ -d "$f" ]]
        then
            run_integration_test "$f"
        fi
    done
}

if [[ -z ${MAKE_TERMOUT+x} ]]
then
    if [[ -z ${LLDAP_CONTAINER_ID+x} ]]
    then
        echo "Starting LLDAP server..."
        start_lldap_server
    else
        echo "Stopping LLDAP server..."
        stop_lldap_server
    fi
else
    run_unit_test
    run_integration_tests
fi
