#!/usr/bin/env bash

# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

set -eo pipefail

readonly DATABASE="postgres"
readonly TEST="$1"

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
        -p 5432 \
        --env "POSTGRES_DB=lldap" \
        --env "POSTGRES_USER=postgres" \
        --env "POSTGRES_PASSWORD=${passwd}" \
        postgres:latest)
    local postgres_port=$(docker inspect --format '{{ (index (index .NetworkSettings.Ports "5432/tcp") 0).HostPort }}' "$postgres_cnt_id")
    local postgres_cnt_ip="127.0.0.1"

    if [[ "$DEBUG" == "true" ]]
    then
        docker logs -f "$postgres_cnt_id" &
    fi

    wait_for_service "$postgres_cnt_ip" "$postgres_port"

    cat <<EOF
export POSTGRES_CONTAINER_ID="$postgres_cnt_id"
export POSTGRES_HOST="$postgres_cnt_ip"
export POSTGRES_PORT="$postgres_port"
export POSTGRES_PASSWORD="$passwd"
EOF
    export POSTGRES_CONTAINER_ID="$postgres_cnt_id"
    export POSTGRES_HOST="$postgres_cnt_ip"
    export POSTGRES_PORT="$postgres_port"
    export POSTGRES_PASSWORD="$passwd"
}

function start_lldap_server {
    local passwd=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 18; echo)
    local lldap_cnt_id
    if [[ "$DATABASE" == "postgres" ]]
    then
        database_url="postgres://postgres:${POSTGRES_PASSWORD}@host.docker.internal:${POSTGRES_PORT}/lldap"
        lldap_cnt_id=$(docker run \
            --detach \
            --rm \
            -p 17170 \
            -p 3890 \
            --add-host=host.docker.internal:host-gateway \
            --env "LLDAP_DATABASE_URL=${database_url}" \
            --env "LLDAP_LDAP_USER_PASS=${passwd}" \
            --env "LLDAP_LDAP_BASE_DN=dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com" \
            --env "LLDAP_JWT_SECRET=$(uuidgen)" \
            lldap/lldap:latest)
    else
        lldap_cnt_id=$(docker run \
            --detach \
            --rm \
            -p 17170 \
            -p 3890 \
            --env "LLDAP_LDAP_USER_PASS=${passwd}" \
            --env "LLDAP_LDAP_BASE_DN=dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com" \
            --env "LLDAP_JWT_SECRET=$(uuidgen)" \
            lldap/lldap:latest)
    fi
    local lldap_port_http=$(docker inspect --format '{{ (index (index .NetworkSettings.Ports "17170/tcp") 0).HostPort }}' "$lldap_cnt_id")
    local lldap_port_ldap=$(docker inspect --format '{{ (index (index .NetworkSettings.Ports "3890/tcp") 0).HostPort }}' "$lldap_cnt_id")
    local lldap_cnt_ip="127.0.0.1"

    if [[ "$DEBUG" == "true" ]]
    then
        docker logs -f "$lldap_cnt_id" &
    fi

    wait_for_service "$lldap_cnt_ip" "$lldap_port_http"

    cat <<EOF
export LLDAP_CONTAINER_ID="$lldap_cnt_id"
export LLDAP_HOST="$lldap_cnt_ip"
export LLDAP_PORT_HTTP="$lldap_port_http"
export LLDAP_PORT_LDAP="$lldap_port_ldap"
export LLDAP_PASSWORD="$passwd"
EOF
    export LLDAP_CONTAINER_ID="$lldap_cnt_id"
    export LLDAP_HOST="$lldap_cnt_ip"
    export LLDAP_PORT_HTTP="$lldap_port_http"
    export LLDAP_PORT_LDAP="$lldap_port_ldap"
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
    export TF_IN_AUTOMATION="yeah"
    cd "$test_path"
    rm -Rvf .terraform .terraform.lock.hcl terraform.tfstate terraform.tfstate.backup
    cat > "$test_path/test.auto.tfvars" << EOF
lldap_http_url="http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}"
lldap_ldap_url="ldap://${LLDAP_HOST}:${LLDAP_PORT_LDAP}"
lldap_username="admin"
lldap_password="$LLDAP_PASSWORD"
lldap_base_dn="dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com"
EOF
    export LLDAP_HTTP_URL="http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}"
    export LLDAP_LDAP_URL="ldap://${LLDAP_HOST}:${LLDAP_PORT_LDAP}"
    export LLDAP_USERNAME="admin"
    export LLDAP_PASSWORD="$LLDAP_PASSWORD"
    tofu init -reconfigure -upgrade
    if [ -e "${test_path}/test.sh" ]
    then
        "${test_path}/test.sh"
    else
        tofu test
    fi
    unset "TF_IN_AUTOMATION"
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
    # macos: arm64, linux: aarch64
    if [[ $(uname -m) == "aarch64" ]] || [[ $(uname -m) == "arm64" ]]
    then
        tf_uname_arch="arm64"
    elif [[ $(uname -m) == "x86_64" ]]
    then
        tf_uname_arch="amd64"
    else
        tf_uname_arch=$(uname -m)
    fi
    tf_uname=$(uname  | tr '[:upper:]' '[:lower:]')
    mkdir -p "${temp_test_dir}/plugins/registry.opentofu.org/tasansga/lldap/0.0.1/${tf_uname}_${tf_uname_arch}/"
    cp "${tf_provider_lldap_root_dir}/dist/terraform-provider-lldap" "${temp_test_dir}/plugins/registry.opentofu.org/tasansga/lldap/0.0.1/${tf_uname}_${tf_uname_arch}/terraform-provider-lldap"

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

    if [ -z "$TEST" ]
    then
        for f in $tf_provider_lldap_root_dir/tests/*
        do
            run_integration_test "$f"
        done
    else
        f="${tf_provider_lldap_root_dir}/tests/${TEST}"
        run_integration_test "$f"
    fi
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
