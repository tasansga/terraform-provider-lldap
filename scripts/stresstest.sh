#!/usr/bin/env bash

set -eo pipefail

function auth() {
  local lldap_http_host="http://${1}:17170"
  local password="$2"
  local response
  response=$(curl \
    --silent \
    --request POST \
    --url "${lldap_http_host}/auth/simple/login" \
    --header 'Content-Type: application/json' \
    --data "{ \"username\" : \"admin\", \"password\" : \"${password}\" }")
  echo "$response" | jq --raw-output '.token'
}

function create_user() {
  local lldap_http_host="$1"
  local token="$2"
  local username="$3"
  local request
  request="$(cat <<EOF
{
  "query" : "mutation CreateUser(\$user: CreateUserInput!) {createUser(user: \$user) {id creationDate}}",
  "operationName" : "CreateUser",
  "variables" : {
    "user" : {
        "id": "${username}",
        "email": "${username}@mail.test"
    }
  }
}
EOF
)"
  local response
  response=$(curl \
    --silent \
    --request POST \
    --url "${lldap_http_host}/api/graphql" \
    --header "Authorization: Bearer ${token}" \
    --header 'Content-Type: application/json' \
    --data "$request")
  if [[ ! "$response" == '{"data":{"createUser":{"id":"testuser'* ]]
  then
    >&2 echo "Unexpected response: '${response}'"
  fi
}

function create_user_and_set_password() {
  local lldap_host="$1"
  local username="$2"
  local admin_password="$3"
  local token=$(auth "$cnt_ip" "$password")
  local password=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 10; echo)
  create_user "http://${lldap_host}:17170" "$token" "testuser${i}"
  ldappasswd \
    -v \
    -H "ldap://${lldap_host}:3890" \
    -D "cn=admin,ou=people,dc=example,dc=com" \
    -w "$admin_password" \
    -s "$password" \
    "cn=testuser${i},ou=people,dc=example,dc=com"
}

password=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 18; echo)
cnt_id=$(docker run \
    --detach \
    --rm \
    --env "LLDAP_LDAP_USER_PASS=${password}" \
    lldap/lldap:latest)

function on_exit {
    docker stop "$cnt_id" > /dev/null || true
}

trap on_exit EXIT

docker logs -f "$cnt_id" &
sleep 3 # need to wait for lldap to properly start up

cnt_ip=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$cnt_id")

for i in $(seq 1 50)
do
  create_user_and_set_password "$cnt_ip" "testuser${i}" "$password" &
done

sleep 20 # need to wait for lldap to process and answer the async requests
