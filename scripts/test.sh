#!/usr/bin/env bash

# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

set -eo pipefail

readonly DATABASE="postgres"
readonly COMMAND="$1"

readonly POSTGRES_VERSION=18.6
readonly LLDAP_VERSION=2026-09-22

function get_container_host_port {
    local cnt_id="$1"
    local container_port="$2"
    local port_line
    for _ in {1..10}; do
        port_line=$(docker port "$cnt_id" "$container_port" 2>/dev/null || true)
        if [[ -n "$port_line" ]]; then
            local first_line="${port_line%%$'\n'*}"
            local host_port="${first_line##*:}"
            echo "$host_port"
            return 0
        fi
        local status
        status=$(docker inspect -f '{{.State.Status}}' "$cnt_id" 2>/dev/null || true)
        if [[ "$status" == "exited" || "$status" == "dead" ]]; then
            echo "Error: Container ${cnt_id} is in status '${status}' while resolving ${container_port}" >&2
            return 1
        fi
        sleep 0.5
    done
    echo "Error: Failed to resolve host port for container ${cnt_id} (${container_port})" >&2
    return 1
}

function wait_for_service {
    local host="$1"
    local port="$2"
    local cnt_id="${3:-}"
    local timeout="${4:-30}"

    if [[ -z "$port" ]]; then
        echo "Error: Service port is empty for host ${host} (container: ${cnt_id:-unknown})" >&2
        if [[ -n "$cnt_id" ]]; then
            docker logs "$cnt_id" 2>&1 || true
        fi
        return 1
    fi

    echo "waiting for ${host}:${port}..."
    local elapsed=0
    while true; do
        if nc -z "$host" "$port" 2>/dev/null; then
            break
        fi
        if [[ -n "$cnt_id" && $(docker inspect -f '{{.State.Running}}' "$cnt_id" 2>/dev/null) != "true" ]]; then
            echo "Error: Container ${cnt_id} exited while waiting for ${host}:${port}" >&2
            docker logs "$cnt_id" 2>&1 || true
            return 1
        fi
        if (( elapsed >= timeout )); then
            echo "Error: Timed out waiting for ${host}:${port} after ${timeout}s" >&2
            if [[ -n "$cnt_id" ]]; then
                docker logs "$cnt_id" 2>&1 || true
            fi
            return 1
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done
}

function start_postgres_server {
    local passwd
    passwd=$(LC_ALL=C tr -dc A-Za-z0-9 </dev/urandom | head -c 18; echo)
    local postgres_cnt_id
    local -a rm_args=(--rm)
    if [[ "$DEBUG" == "true" ]]; then
        rm_args=()
    fi
    postgres_cnt_id=$(docker run \
        --detach \
        "${rm_args[@]}" \
        -p 5432 \
        --env "POSTGRES_DB=lldap" \
        --env "POSTGRES_USER=postgres" \
        --env "POSTGRES_PASSWORD=${passwd}" \
        "docker.io/postgres:$POSTGRES_VERSION")
    local postgres_port
    postgres_port=$(get_container_host_port "$postgres_cnt_id" "5432/tcp") || return 1
    local postgres_cnt_ip="127.0.0.1"

    if [[ "$DEBUG" == "true" ]]
    then
        docker logs -f "$postgres_cnt_id" &
    fi

    echo "waiting for postgres to accept connections..."
    local elapsed=0
    while true; do
        if docker exec "$postgres_cnt_id" pg_isready -h 127.0.0.1 -p 5432 -U postgres -d lldap 2>/dev/null; then
            break
        fi
        if [[ $(docker inspect -f '{{.State.Running}}' "$postgres_cnt_id" 2>/dev/null) != "true" ]]; then
            echo "Error: Postgres container ${postgres_cnt_id} exited prematurely" >&2
            docker logs "$postgres_cnt_id" 2>&1 || true
            return 1
        fi
        if (( elapsed >= 30 )); then
            echo "Error: Timed out waiting for postgres after 30s" >&2
            docker logs "$postgres_cnt_id" 2>&1 || true
            return 1
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done

    wait_for_service "$postgres_cnt_ip" "$postgres_port" "$postgres_cnt_id"

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
    local passwd
    passwd=$(LC_ALL=C tr -dc A-Za-z0-9 </dev/urandom | head -c 18; echo)
    local lldap_cnt_id
    local database_url
    local -a rm_args=(--rm)
    if [[ "$DEBUG" == "true" ]]; then
        rm_args=()
    fi
    if [[ "$DATABASE" == "postgres" ]]
    then
        database_url="postgres://postgres:${POSTGRES_PASSWORD}@host.docker.internal:${POSTGRES_PORT}/lldap"
        lldap_cnt_id=$(docker run \
            --detach \
            "${rm_args[@]}" \
            -p 17170 \
            -p 3890 \
            --add-host=host.docker.internal:host-gateway \
            --env "LLDAP_DATABASE_URL=${database_url}" \
            --env "LLDAP_LDAP_USER_PASS=${passwd}" \
            --env "LLDAP_LDAP_BASE_DN=dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com" \
            --env "LLDAP_JWT_SECRET=$(uuidgen)" \
            "docker.io/lldap/lldap:$LLDAP_VERSION")
    else
        lldap_cnt_id=$(docker run \
            --detach \
            "${rm_args[@]}" \
            -p 17170 \
            -p 3890 \
            --env "LLDAP_LDAP_USER_PASS=${passwd}" \
            --env "LLDAP_LDAP_BASE_DN=dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com" \
            --env "LLDAP_JWT_SECRET=$(uuidgen)" \
            "docker.io/lldap/lldap:$LLDAP_VERSION")
    fi

    if [[ "$DEBUG" == "true" ]]
    then
        docker logs -f "$lldap_cnt_id" &
    fi

    local lldap_port_http
    lldap_port_http=$(get_container_host_port "$lldap_cnt_id" "17170/tcp") || return 1
    local lldap_port_ldap
    lldap_port_ldap=$(get_container_host_port "$lldap_cnt_id" "3890/tcp") || return 1
    local lldap_cnt_ip="127.0.0.1"

    wait_for_service "$lldap_cnt_ip" "$lldap_port_http" "$lldap_cnt_id"
    wait_for_service "$lldap_cnt_ip" "$lldap_port_ldap" "$lldap_cnt_id"

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
    export LLDAP_BASE_DN="dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com"
}

function start_server {
    if [[ "$DATABASE" == "postgres" ]]
    then
        start_postgres_server
    fi
    start_lldap_server
}

function stop_postgres_server {
    if [[ -n "${POSTGRES_CONTAINER_ID:-}" ]]; then
        docker stop "$POSTGRES_CONTAINER_ID" || true
    fi
    cat <<EOF
unset POSTGRES_CONTAINER_ID
unset POSTGRES_HOST
unset POSTGRES_PORT
unset POSTGRES_PASSWORD
EOF
    unset POSTGRES_CONTAINER_ID
    unset POSTGRES_HOST
    unset POSTGRES_PORT
    unset POSTGRES_PASSWORD
}

function stop_lldap_server {
    if [[ -n "${LLDAP_CONTAINER_ID:-}" ]]; then
        docker stop "$LLDAP_CONTAINER_ID" || true
    fi
    cat <<EOF
unset LLDAP_CONTAINER_ID
unset LLDAP_HOST
unset LLDAP_PORT_HTTP
unset LLDAP_PORT_LDAP
unset LLDAP_PASSWORD
unset LLDAP_BASE_DN
EOF
    unset LLDAP_CONTAINER_ID
    unset LLDAP_HOST
    unset LLDAP_PORT_HTTP
    unset LLDAP_PORT_LDAP
    unset LLDAP_PASSWORD
    unset LLDAP_BASE_DN
}

function stop_server {
    if [[ "$DATABASE" == "postgres" ]]
    then
        stop_postgres_server
    fi
    stop_lldap_server
}

function run_unit_test_cli {
    echo "Running CLI unit tests..."
    go test -v ./cmd/lldap-cli
}

function run_unit_test {
    echo "Running unit tests..."
    go test -v ./lldap
}

function run_integration_test_cli {
    echo "Running CLI integration tests..."
    trap stop_server RETURN
    trap stop_server EXIT
    start_server

    LLDAP_BASE_DN="dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com" \
        LLDAP_HTTP_URL="http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}" \
        LLDAP_LDAP_URL="ldap://${LLDAP_HOST}:${LLDAP_PORT_LDAP}" \
        LLDAP_PASSWORD="$LLDAP_PASSWORD" \
        go test -tags=integration -v ./cmd/lldap-cli
}

function run_integration_test_lldap {
    echo "Running lldap package integration tests..."
    trap stop_server RETURN
    trap stop_server EXIT
    start_server

    go test -tags=integration -v ./lldap
}

function run_integration_test {
    local input_path="$1"
    local root_dir="${2:-}"
    local test_path=""
    local test_name=""

    # Handle different input formats
    if [[ "$input_path" == */* ]] && [[ "$input_path" != */tests/* ]]; then
        # Input is like "env/basic" - extract test name and build full path
        test_name="${input_path##*/}"
        local dir_name="${input_path%/*}"
        if [[ -z "$root_dir" ]]; then
            root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
        fi
        test_path="${root_dir}/tests/${dir_name}"
    else
        # Input is a full path like "/path/to/tests/env"
        test_path="$input_path"
    fi

    trap stop_server RETURN
    trap stop_server EXIT
    start_server

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
    export LLDAP_USER="admin"
    export LLDAP_PASSWORD="$LLDAP_PASSWORD"
    export LLDAP_BASE_DN="dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com"
    tofu init -reconfigure -upgrade

    if [ -n "$test_name" ] && [ -e "${test_path}/test_${test_name}.sh" ]; then
        # Run specific test file
        "${test_path}/test_${test_name}.sh"
    elif [ -n "$test_name" ]; then
        echo "Test file test_${test_name}.sh not found in $test_path"
        exit 1
    elif [ -e "${test_path}/test.sh" ]; then
        # Run main test.sh if no specific test requested
        "${test_path}/test.sh"
    else
        # Run all test_*.sh files in the directory
        local -a test_files=("${test_path}"/test_*.sh)
        if [[ -e "${test_files[0]}" ]]; then
            for test_file in "${test_files[@]}"; do
                echo "Running $(basename "$test_file")..."
                "$test_file"
            done
        else
            # Fallback to tofu test
            tofu test
        fi
    fi
    unset "TF_IN_AUTOMATION"
}

function run_integration_tests {
    local scripts_dir
    scripts_dir=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)
    local tf_provider_lldap_root_dir
    tf_provider_lldap_root_dir=$(realpath "${scripts_dir}/..")
    local temp_test_dir
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
    export LLDAP_CLI="${tf_provider_lldap_root_dir}/dist/lldap-cli"
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
        for f in "$tf_provider_lldap_root_dir"/tests/*
        do
            run_integration_test "$f" "$tf_provider_lldap_root_dir"
        done
    else
        # Handle both "env" and "env/basic" formats
        if [[ "$TEST" == */* ]]; then
            # TEST is like "env/basic" - pass the full path
            run_integration_test "$TEST" "$tf_provider_lldap_root_dir"
        else
            # TEST is like "env" - pass just the directory
            local f="${tf_provider_lldap_root_dir}/tests/${TEST}"
            run_integration_test "$f" "$tf_provider_lldap_root_dir"
        fi
    fi
}

case $COMMAND in
    unittest)
        run_unit_test
        ;;
    unittest-cli)
        run_unit_test_cli
        ;;
    inttest)
        run_integration_tests
        ;;
    inttest-cli)
        run_integration_test_cli
        ;;
    inttest-lldap)
        run_integration_test_lldap
        ;;
    *)
        echo "Invalid option. Use: unittest, unittest-cli, inttest, inttest-cli, inttest-lldap, or all."
        exit 1
        ;;
esac
