#!/usr/bin/env bash

# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

set -eo pipefail

readonly DATABASE="postgres"

function wait_for_service {
    local host="$1"
    local port="$2"
    echo "waiting for ${host}:${port}..."
    while true
    do
        sleep 1
        nc -z "$host" "$port" && break
    done
}

function start_postgres_server {
    local passwd=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 18; echo)
    postgres_cnt_id=$(docker run \
        --detach \
        --rm \
        --env "POSTGRES_DB=lldap" \
        --env "POSTGRES_USER=postgres" \
        --env "POSTGRES_PASSWORD=${passwd}" \
        postgres:latest)
    local postgres_cnt_ip=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$postgres_cnt_id")

    if [[ "$DEBUG" == "true" ]]
    then
        docker logs -f "$postgres_cnt_id" &
    fi

    wait_for_service "$postgres_cnt_ip" 5432

    cat <<EOF
export POSTGRES_CONTAINER_ID="$postgres_cnt_id"
export POSTGRES_HOST="$postgres_cnt_ip"
export POSTGRES_PASSWORD="$passwd"
EOF
    export POSTGRES_CONTAINER_ID="$postgres_cnt_id"
    export POSTGRES_HOST="$postgres_cnt_ip"
    export POSTGRES_PASSWORD="$passwd"
}

function start_lldap_server {
    local passwd=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 18; echo)
    local lldap_cnt_id
    if [[ "$DATABASE" == "postgres" ]]
    then
        database_url="postgres://postgres:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/lldap"
        lldap_cnt_id=$(docker run \
            --detach \
            --rm \
            --env "LLDAP_DATABASE_URL=${database_url}" \
            --env "LLDAP_LDAP_USER_PASS=${passwd}" \
            --env "LLDAP_LDAP_BASE_DN=dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com" \
            lldap/lldap:latest)
    else
        lldap_cnt_id=$(docker run \
            --detach \
            --rm \
            --env "LLDAP_LDAP_USER_PASS=${passwd}" \
            --env "LLDAP_LDAP_BASE_DN=dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com" \
            lldap/lldap:latest)
    fi
    local lldap_cnt_ip=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$lldap_cnt_id")

    if [[ "$DEBUG" == "true" ]]
    then
        docker logs -f "$lldap_cnt_id" &
    fi

    wait_for_service "$lldap_cnt_ip" 17170

    cat <<EOF
export LLDAP_CONTAINER_ID="$lldap_cnt_id"
export LLDAP_HOST="$lldap_cnt_ip"
export LLDAP_PASSWORD="$passwd"
EOF
    export LLDAP_CONTAINER_ID="$lldap_cnt_id"
    export LLDAP_HOST="$lldap_cnt_ip"
    export LLDAP_PASSWORD="$passwd"
}

function start_server {
    if [[ "$DATABASE" == "postgres" ]]
    then
        start_postgres_server
    fi
    start_lldap_server
}

function stop_postgres_server {
    docker stop "$POSTGRES_CONTAINER_ID" || true
    cat <<EOF
unset POSTGRES_CONTAINER_ID
unset POSTGRES_HOST
unset POSTGRES_PASSWORD
EOF
    unset POSTGRES_CONTAINER_ID
    unset POSTGRES_HOST
    unset POSTGRES_PASSWORD
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

function stop_server {
    if [[ "$DATABASE" == "postgres" ]]
    then
        stop_postgres_server
    fi
    stop_lldap_server
}

function run_unit_test {
    start_server
    trap stop_server RETURN
    trap stop_server EXIT

    go test ./lldap
}

function run_integration_test {
    local test_path="$1"
    start_server
    trap stop_server RETURN
    trap stop_server EXIT

    echo "Running test: ${test_path}"
    cd "$test_path"
    rm -Rvf .terraform .terraform.lock.hcl terraform.tfstate terraform.tfstate.backup
    cat > "$test_path/.tfvars" << EOF
lldap_http_url="http://${LLDAP_HOST}:17170"
lldap_ldap_url="ldap://${LLDAP_HOST}:3890"
lldap_username="admin"
lldap_password="$LLDAP_PASSWORD"
lldap_base_dn="dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com"
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
    cp "${tf_provider_lldap_root_dir}/dist/terraform-provider-lldap" "${temp_test_dir}/plugins/registry.opentofu.org/tasansga/lldap/0.0.1/linux_amd64/terraform-provider-lldap"

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
        echo "Starting LLDAP server... (set DEBUG=true for logs)"
        start_server
    else
        echo "Stopping LLDAP server..."
        stop_server
    fi
else
    run_unit_test
    run_integration_tests
fi
