#!/usr/bin/env bash

set -exo pipefail

echo "=== Basic Group Lifecycle Test ==="
tofu apply -auto-approve
tofu apply -auto-approve -destroy

echo "=== Test Group Count Scaling ==="
# Start with 0 groups
tofu apply -auto-approve -var group_count=0

# Scale up to 3 groups (changes expected)
tofu apply -auto-approve -var group_count=3

# Scale up to 10 groups (changes expected)
tofu apply -auto-approve -var group_count=10

# Scale down to 5 groups (changes expected)
tofu apply -auto-approve -var group_count=5

# Scale down to 1 group (changes expected)
tofu apply -auto-approve -var group_count=1

# Clean up
tofu apply -auto-approve -destroy -var group_count=1

echo "=== Test Group with Attributes Lifecycle ==="
# Create groups without attributes
tofu apply -auto-approve -var create_group_with_attrs=false

# Add attributes
tofu apply -auto-approve -var create_group_with_attrs=true

# Remove attributes
tofu apply -auto-approve -var create_group_with_attrs=false

tofu apply -auto-approve -destroy -var create_group_with_attrs=false

echo "=== Test Group Memberships Lifecycle ==="
# Create groups and users without memberships
tofu apply -auto-approve -var create_group_with_members=false

# Add memberships
tofu apply -auto-approve -var create_group_with_members=true

# Remove users (should remove memberships)
tofu apply -auto-approve -var create_test_users=false -var create_group_with_members=false

tofu apply -auto-approve -destroy -var create_test_users=false -var create_group_with_members=false

echo "=== Test Complex State Transitions ==="
# Start with everything
tofu apply -auto-approve -var group_count=3 -var create_group_with_attrs=true -var create_test_users=true -var create_group_with_members=true

# Remove memberships only
tofu apply -auto-approve -var group_count=3 -var create_group_with_attrs=true -var create_test_users=true -var create_group_with_members=false

# Remove attributes
tofu apply -auto-approve -var group_count=3 -var create_group_with_attrs=false -var create_test_users=true -var create_group_with_members=false

# Scale down groups
tofu apply -auto-approve -var group_count=1 -var create_group_with_attrs=false -var create_test_users=true -var create_group_with_members=false

# Remove users
tofu apply -auto-approve -var group_count=1 -var create_group_with_attrs=false -var create_test_users=false -var create_group_with_members=false

tofu apply -auto-approve -destroy

echo "=== Test Random Updates ==="
tofu apply -auto-approve
# Update random suffix to trigger name changes
tofu taint random_string.suffix
tofu apply -auto-approve
tofu apply -auto-approve -destroy

echo "=== Test Edge Cases ==="
# Test with 0 groups
tofu apply -auto-approve -var group_count=0 -var create_group_with_attrs=false -var create_test_users=false -var create_group_with_members=false
tofu apply -auto-approve -destroy

# Test with maximum groups
tofu apply -auto-approve -var group_count=20
tofu apply -auto-approve -destroy -var group_count=20

echo "=== Out-of-Band Group Changes Test ==="
tofu apply -auto-approve -var group_count=2 -var create_group_with_attrs=true -var create_test_users=true -var create_group_with_members=true

# Get the created resources for out-of-band testing
GROUP_IDS=$(tofu output -json created_groups | jq -r '.[].id')
FIRST_GROUP_ID=$(echo "$GROUP_IDS" | head -n1)
USER_ID=$(tofu output -json test_users | jq -r '.[0].username')

# Set environment variables for lldap-cli
export LLDAP_BASE_DN="dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com"
export LLDAP_HTTP_URL="http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}"
export LLDAP_LDAP_URL="ldap://${LLDAP_HOST}:${LLDAP_PORT_LDAP}"
export LLDAP_USER="admin"

echo "Testing out-of-band group display name change..."
# Change group display name out-of-band
../../dist/lldap-cli group update "$FIRST_GROUP_ID" --displayname "Hacked Group Name"

# Terraform should detect and fix the drift
echo "Checking if Terraform detects the group drift..."
tofu plan  # Should show changes to revert display name
echo "Applying to fix the group drift..."
tofu apply -auto-approve  # Should fix the drift

echo "Testing out-of-band group deletion..."
# Delete a group out-of-band
../../dist/lldap-cli group delete "$FIRST_GROUP_ID"

# Terraform should detect and recreate the group
echo "Checking if Terraform detects the group deletion..."
tofu plan  # Should show group recreation
echo "Applying to recreate the group..."
tofu apply -auto-approve  # Should recreate the group

echo "Testing out-of-band user removal from group..."
# Remove user from group out-of-band (if memberships exist)
if [ -n "$USER_ID" ]; then
    ../../dist/lldap-cli member remove "$FIRST_GROUP_ID" "$USER_ID" || echo "User not in group or command failed"

    # Terraform should detect and fix the membership
    echo "Checking if Terraform detects the membership change..."
    tofu plan  # Should show membership recreation
    echo "Applying to fix the membership..."
    tofu apply -auto-approve  # Should fix the membership
fi

tofu apply -auto-approve -destroy

echo "=== All group lifecycle tests completed successfully! ==="
