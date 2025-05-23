#!/usr/bin/env bash

set -exo pipefail

echo "=== Basic Lifecycle Test: Apply, Plan, Destroy ==="
tofu apply -auto-approve
tofu apply -auto-approve -destroy

echo "=== Test Data Source Consistency ==="
tofu apply -auto-approve
# Verify data sources return consistent results
tofu refresh
tofu apply -auto-approve -destroy

echo "=== Test Resource Updates ==="
tofu apply -auto-approve
# Update user display name by tainting random string
tofu taint random_string.suffix
tofu apply -auto-approve
tofu apply -auto-approve -destroy

echo "=== Test Password Updates ==="
tofu apply -auto-approve
# Update password
tofu taint random_password.user_password
tofu apply -auto-approve
tofu apply -auto-approve -destroy

echo "=== Test Conditional Resource Creation ==="
# Test without user attribute
tofu apply -auto-approve -var create_user_attr=false
# Add user attribute
tofu apply -auto-approve -var create_user_attr=true
# Remove user attribute
tofu apply -auto-approve -var create_user_attr=false
tofu apply -auto-approve -destroy -var create_user_attr=false

echo "=== Test Group Attribute Lifecycle ==="
# Test without group attribute
tofu apply -auto-approve -var create_group_attr=false
# Add group attribute
tofu apply -auto-approve -var create_group_attr=true
# Remove group attribute
tofu apply -auto-approve -var create_group_attr=false
tofu apply -auto-approve -destroy -var create_group_attr=false

echo "=== Test Complex State Changes ==="
# Create everything
tofu apply -auto-approve
# Remove user attribute only
tofu apply -auto-approve -var create_user_attr=false
# Remove group attribute only
tofu apply -auto-approve -var create_user_attr=false -var create_group_attr=false
# Add both back
tofu apply -auto-approve -var create_user_attr=true -var create_group_attr=true
tofu plan -var create_user_attr=true -var create_group_attr=true  # Check for any infrastructure changes
tofu apply -auto-approve -destroy

echo "=== Out-of-Band Changes Test with lldap-cli ==="
tofu apply -auto-approve

# Get the created resources for out-of-band testing
USER_ID=$(tofu output -json test_user | jq -r '.username')
GROUP_ID=$(tofu output -json test_group | jq -r '.id')

echo "Testing out-of-band user changes..."
# Set environment variables for lldap-cli
export LLDAP_BASE_DN="dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com"
export LLDAP_HTTP_URL="http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}"
export LLDAP_LDAP_URL="ldap://${LLDAP_HOST}:${LLDAP_PORT_LDAP}"
export LLDAP_USER="admin"

# Change user email out-of-band
../../dist/lldap-cli user update "$USER_ID" --email "changed@example.com"

# Terraform should detect and fix the drift
echo "Checking if Terraform detects the drift..."
tofu plan  # Should show changes to revert email
echo "Applying to fix the drift..."
tofu apply -auto-approve  # Should fix the drift

echo "Testing out-of-band group changes..."
# Change group display name out-of-band
../../dist/lldap-cli group update "$GROUP_ID" --displayname "Changed Group Name"

# Terraform should detect and fix the drift
echo "Checking if Terraform detects the group drift..."
tofu plan  # Should show changes to revert display name
echo "Applying to fix the group drift..."
tofu apply -auto-approve  # Should fix the drift

echo "Testing out-of-band user deletion..."
# Delete user out-of-band
../../dist/lldap-cli user delete "$USER_ID"

# Terraform should detect and recreate the user
echo "Checking if Terraform detects the user deletion..."
tofu plan  # Should show user recreation
echo "Applying to recreate the user..."
tofu apply -auto-approve  # Should recreate the user

tofu apply -auto-approve -destroy

echo "=== All basic lifecycle tests completed successfully! ==="
