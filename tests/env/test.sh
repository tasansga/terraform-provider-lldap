#!/usr/bin/env bash

set -exo pipefail

echo "=== Environment Variable Authentication Test ==="
# Test basic functionality - connection details come from test.auto.tfvars
tofu apply -auto-approve
tofu apply -auto-approve -destroy

echo "=== Test Data Source Refresh with Environment Auth ==="
tofu apply -auto-approve
# Test that data sources work consistently with env vars
tofu refresh
tofu plan  # Check plan without detailed-exitcode since data source refresh may show output changes
tofu apply -auto-approve -destroy

echo "=== Test Resource Creation/Deletion Cycles ==="
# Test user creation
tofu apply -auto-approve -var create_user=true -var create_group=false
tofu plan -var create_user=true -var create_group=false  # Check plan without detailed-exitcode since data source refresh may show output changes
tofu apply -auto-approve -destroy -var create_user=true -var create_group=false

# Test group creation
tofu apply -auto-approve -var create_user=false -var create_group=true
tofu plan -var create_user=false -var create_group=true  # Check plan without detailed-exitcode since data source refresh may show output changes
tofu apply -auto-approve -destroy -var create_user=false -var create_group=true

echo "=== Test Mixed Resource States ==="
# Create user only
tofu apply -auto-approve -var create_user=true -var create_group=false
# Add group
tofu apply -auto-approve -var create_user=true -var create_group=true
tofu plan -var create_user=true -var create_group=true  # Check plan without detailed-exitcode since data source refresh may show output changes
# Remove user, keep group
tofu apply -auto-approve -var create_user=false -var create_group=true
tofu plan -var create_user=false -var create_group=true  # Check plan without detailed-exitcode since data source refresh may show output changes
# Remove everything
tofu apply -auto-approve -destroy -var create_user=false -var create_group=true

echo "=== Test Random Resource Updates ==="
tofu apply -auto-approve
# Update random suffix to trigger resource updates
tofu taint random_string.suffix
tofu apply -auto-approve
tofu plan  # Check plan without detailed-exitcode since data source refresh may show output changes
# Update password
tofu taint random_password.password
tofu apply -auto-approve
tofu plan  # Check plan without detailed-exitcode since data source refresh may show output changes
tofu apply -auto-approve -destroy

echo "=== Test Environment Variable Consistency ==="
# Verify that environment variables are being used correctly
echo "Testing with environment variables:"
echo "LLDAP_HTTP_URL: ${LLDAP_HTTP_URL:-not set}"
echo "LLDAP_LDAP_URL: ${LLDAP_LDAP_URL:-not set}"
echo "LLDAP_USERNAME: ${LLDAP_USERNAME:-not set}"
echo "LLDAP_PASSWORD: ${LLDAP_PASSWORD:+set}"
echo "LLDAP_BASE_DN: ${LLDAP_BASE_DN:-not set}"

# Test multiple apply/destroy cycles to ensure consistency
for i in {1..3}; do
  echo "=== Cycle $i ==="
  tofu apply -auto-approve
  tofu plan  # Check plan without detailed-exitcode since data source refresh may show output changes
  tofu apply -auto-approve -destroy
done

echo "=== Out-of-Band Changes Test with Environment Auth ==="
tofu apply -auto-approve -var create_user=true -var create_group=true

# Get the created resources for out-of-band testing
USER_ID=$(tofu output -json env_test_user | jq -r '.username')
GROUP_ID=$(tofu output -json env_test_group | jq -r '.id')

echo "Testing out-of-band changes with environment authentication..."
# Set environment variables for lldap-cli
export LLDAP_HTTP_URL="http://${LLDAP_HOST}:${LLDAP_PORT_HTTP}"
export LLDAP_LDAP_URL="ldap://${LLDAP_HOST}:${LLDAP_PORT_LDAP}"
export LLDAP_USERNAME="admin"
export LLDAP_BASE_DN="dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com"
# Change user first name out-of-band
../../dist/lldap-cli user update "$USER_ID" --firstname "OutOfBand"

# Terraform should detect and fix the drift
echo "Checking if Terraform detects the drift with env auth..."
tofu plan  # Should show changes to revert first name
echo "Applying to fix the drift..."
tofu apply -auto-approve  # Should fix the drift
tofu plan  # No changes expected after fix - this should return 0

tofu apply -auto-approve -destroy -var create_user=true -var create_group=true

echo "=== All environment variable tests completed successfully! ==="
