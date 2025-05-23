terraform {
  required_providers {
    lldap = {
      source  = "tasansga/lldap"
      version = "0.0.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.3"
    }
  }
}

variable "lldap_http_url" {}
variable "lldap_ldap_url" {}
variable "lldap_username" {}
variable "lldap_password" {}
variable "lldap_base_dn" {}

provider "lldap" {
  http_url = var.lldap_http_url
  ldap_url = var.lldap_ldap_url
  username = var.lldap_username
  password = var.lldap_password
  base_dn  = var.lldap_base_dn
}

# Variable for controlling group count
variable "group_count" {
  type    = number
  default = 5
  validation {
    condition     = var.group_count >= 0 && var.group_count <= 50
    error_message = "group_count must be between 0 and 50."
  }
}

# Random suffix for unique naming
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

# Create multiple groups for lifecycle testing
resource "lldap_group" "test_groups" {
  count        = var.group_count
  display_name = "Test Group ${count.index} - ${random_string.suffix.result}"
}

# Test group with custom attributes
variable "create_group_with_attrs" {
  type    = bool
  default = true
}

resource "lldap_group_attribute" "test_attr" {
  count          = var.create_group_with_attrs ? 1 : 0
  name           = "testgroupattr${random_string.suffix.result}"
  attribute_type = "STRING"
  is_list        = false
  is_visible     = true
}

resource "lldap_group" "group_with_attrs" {
  count        = var.create_group_with_attrs ? 1 : 0
  display_name = "Group with Attributes - ${random_string.suffix.result}"
  
  depends_on = [lldap_group_attribute.test_attr]
}

# Test users for group membership testing
variable "create_test_users" {
  type    = bool
  default = true
}

resource "random_password" "user_passwords" {
  count  = var.create_test_users ? 3 : 0
  length = 16
}

resource "lldap_user" "test_users" {
  count        = var.create_test_users ? 3 : 0
  username     = "testuser${count.index}-${random_string.suffix.result}"
  email        = "testuser${count.index}-${random_string.suffix.result}@group.test"
  password     = random_password.user_passwords[count.index].result
  display_name = "Test User ${count.index}"
  first_name   = "Test"
  last_name    = "User${count.index}"
}

# Test group with members
variable "create_group_with_members" {
  type    = bool
  default = true
}

resource "lldap_group" "group_with_members" {
  count        = var.create_group_with_members && var.create_test_users ? 1 : 0
  display_name = "Group with Members - ${random_string.suffix.result}"
}

# Test group memberships
resource "lldap_group_memberships" "test_memberships" {
  count    = var.create_group_with_members && var.create_test_users ? 1 : 0
  group_id = lldap_group.group_with_members[0].id
  user_ids = [for user in lldap_user.test_users : user.id]
}

# Data sources for verification
data "lldap_groups" "all_groups" {
  depends_on = [lldap_group.test_groups, lldap_group.group_with_attrs, lldap_group.group_with_members]
}

data "lldap_group_attributes" "all_group_attrs" {
  depends_on = [lldap_group_attribute.test_attr]
}

# Outputs for verification
output "group_count" {
  value = var.group_count
}

output "created_groups" {
  value = lldap_group.test_groups
}

output "all_groups" {
  value = data.lldap_groups.all_groups.groups
}

output "group_with_attrs" {
  value = var.create_group_with_attrs ? lldap_group.group_with_attrs[0] : null
}

output "test_attr" {
  value = var.create_group_with_attrs ? lldap_group_attribute.test_attr[0] : null
}

output "group_with_members" {
  value = var.create_group_with_members && var.create_test_users ? lldap_group.group_with_members[0] : null
}

output "test_users" {
  value     = var.create_test_users ? lldap_user.test_users : []
  sensitive = true
}

output "group_memberships" {
  value = var.create_group_with_members && var.create_test_users ? lldap_group_memberships.test_memberships[0] : null
}

output "all_group_attrs" {
  value = data.lldap_group_attributes.all_group_attrs
}
