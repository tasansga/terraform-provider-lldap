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

# Variables for LLDAP connection (matching test script format)
variable "lldap_http_url" {}
variable "lldap_ldap_url" {}
variable "lldap_username" {}
variable "lldap_password" {}
variable "lldap_base_dn" {}

# Variables for testing different configurations
variable "test_attribute_type" {
  description = "Attribute type for testing"
  type        = string
  default     = "STRING"
}

variable "test_is_list" {
  description = "Whether the attribute is a list"
  type        = bool
  default     = false
}

variable "test_values" {
  description = "Values to assign to the attribute"
  type        = list(string)
  default     = ["test-value-1"]
}

provider "lldap" {
  http_url = var.lldap_http_url
  ldap_url = var.lldap_ldap_url
  username = var.lldap_username
  password = var.lldap_password
  base_dn  = var.lldap_base_dn
}

# Random suffix for unique resource names
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

# Create a test group
resource "lldap_group" "test_group" {
  display_name = "Test Group ${random_string.suffix.result}"
}

# Create a custom group attribute
resource "lldap_group_attribute" "test_attr" {
  name           = "testgroupattr"
  attribute_type = var.test_attribute_type
  is_list        = var.test_is_list
  is_visible     = true
}

# Assign values to the custom attribute for the group
resource "lldap_group_attribute_assignment" "test_assignment" {
  group_id     = lldap_group.test_group.id
  attribute_id = lldap_group_attribute.test_attr.id
  value        = var.test_attribute_type == "JPEG_PHOTO" ? [filebase64("${path.module}/test.jpeg")] : var.test_values
}

# Data sources for validation
data "lldap_group" "test_group" {
  id = lldap_group.test_group.id
  depends_on = [lldap_group_attribute_assignment.test_assignment]
}

# Outputs for testing
output "test_group" {
  value = lldap_group.test_group
}

output "test_group_attr" {
  value = lldap_group_attribute.test_attr
}

output "test_assignment" {
  value = lldap_group_attribute_assignment.test_assignment
}

output "group_with_attributes" {
  value = data.lldap_group.test_group
}
