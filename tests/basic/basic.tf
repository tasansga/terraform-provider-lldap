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

# Test basic data sources
data "lldap_users" "users" {}
data "lldap_groups" "groups" {}
data "lldap_user_attributes" "user_attrs" {}
data "lldap_group_attributes" "group_attrs" {}

# Create test resources for lifecycle testing
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

resource "random_password" "user_password" {
  length = 16
}

# Test user lifecycle
resource "lldap_user" "test_user" {
  username     = "testuser-${random_string.suffix.result}"
  email        = "testuser-${random_string.suffix.result}@example.com"
  password     = random_password.user_password.result
  display_name = "Test User ${random_string.suffix.result}"
  first_name   = "Test"
  last_name    = "User"
}

# Test group lifecycle
resource "lldap_group" "test_group" {
  display_name = "Test Group ${random_string.suffix.result}"
}

# Test user attribute lifecycle
variable "create_user_attr" {
  type    = bool
  default = true
}

resource "lldap_user_attribute" "test_attr" {
  count          = var.create_user_attr ? 1 : 0
  name           = "test-attr-${random_string.suffix.result}"
  attribute_type = "STRING"
  is_list        = false
  is_visible     = true
  is_editable    = true
}

# Test group attribute lifecycle
variable "create_group_attr" {
  type    = bool
  default = true
}

resource "lldap_group_attribute" "test_attr" {
  count          = var.create_group_attr ? 1 : 0
  name           = "test-group-attr-${random_string.suffix.result}"
  attribute_type = "STRING"
  is_list        = false
  is_visible     = true
}

# Outputs for verification
output "users" {
  value = data.lldap_users.users.users
}

output "groups" {
  value = data.lldap_groups.groups.groups
}

output "user_attrs" {
  value = data.lldap_user_attributes.user_attrs
}

output "group_attrs" {
  value = data.lldap_group_attributes.group_attrs
}

output "test_user" {
  value     = lldap_user.test_user
  sensitive = true
}

output "test_group" {
  value = lldap_group.test_group
}

output "test_user_attr" {
  value = var.create_user_attr ? lldap_user_attribute.test_attr[0] : null
}

output "test_group_attr" {
  value = var.create_group_attr ? lldap_group_attribute.test_attr[0] : null
}
