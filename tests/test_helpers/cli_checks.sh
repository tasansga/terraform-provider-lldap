#!/usr/bin/env bash

# Provides a quick verification that lldap-cli sees the resources terraform
# touches before the tests clean them up.

set -uo pipefail

if [ -n "${LLDAP_CLI:-}" ]; then
  LLDAP_CLI_BIN="$LLDAP_CLI"
else
  helper_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  repo_root="$(cd "$helper_dir/../.." && pwd)"
  LLDAP_CLI_BIN="$repo_root/dist/lldap-cli"
fi

CLI_OUTPUT_CHECKS=(
  "user:test_user"
  "group:test_group"
  "user:env_test_user"
  "group:env_test_group"
  "group-first:created_groups"
  "user-first:created_users"
  "group-first:test_groups"
  "user-first:test_users"
  "group-first:all_groups"
  "user-first:all_users"
  "user:main_user"
  "user:nopasswd_user"
  "user:user_with_attrs"
  "group:group_with_attrs"
  "group:group_with_members"
  "group:test_group_detailed"
  "user-first:users"
  "group-first:groups"
  "group-attr:test_group_attr"
  "user-attr:test_user_attr"
  "user-attr:test_attr"
)

run_cli_check() {
  local kind="$1"
  local payload="$2"
  local identifier
  case "$kind" in
    user)
      identifier=$(printf '%s' "$payload" | jq -r '(.username // .id // empty)')
      ;;
    user-first)
      identifier=$(printf '%s' "$payload" | jq -r 'if type == "array" and length > 0 then .[0] | (.username // .id // empty) else "" end')
      ;;
    group)
      identifier=$(printf '%s' "$payload" | jq -r '(.id // empty)')
      ;;
    group-first)
      identifier=$(printf '%s' "$payload" | jq -r 'if type == "array" and length > 0 then .[0].id // empty else "" end')
      ;;
    group-attr)
      identifier=$(printf '%s' "$payload" | jq -r '(.name // empty)')
      ;;
    user-attr)
      identifier=$(printf '%s' "$payload" | jq -r '(.name // empty)')
      ;;
    *)
      return 0
      ;;
  esac

  if [ -z "$identifier" ] || [ "$identifier" = "null" ]; then
    return 0
  fi

  case "$kind" in
    user|user-first)
      "$LLDAP_CLI_BIN" user get "$identifier" >/dev/null
      ;;
    group|group-first)
      "$LLDAP_CLI_BIN" group get "$identifier" >/dev/null
      ;;
    group-attr)
      "$LLDAP_CLI_BIN" attribute schema --group "$identifier" >/dev/null
      ;;
    user-attr)
      "$LLDAP_CLI_BIN" attribute schema --user "$identifier" >/dev/null
      ;;
  esac
}

cli_assert_state_change() {
  if [ ! -x "$LLDAP_CLI_BIN" ]; then
    echo "lldap-cli not found at $LLDAP_CLI_BIN" >&2
    return 1
  fi

  local check
  local raw
  for check in "${CLI_OUTPUT_CHECKS[@]}"; do
    local kind="${check%%:*}"
    local output="${check#*:}"
    if ! raw=$(tofu output -json "$output" 2>/dev/null); then
      continue
    fi
    if [ -z "$raw" ] || [ "$raw" = "null" ]; then
      continue
    fi
    run_cli_check "$kind" "$raw"
  done
}
