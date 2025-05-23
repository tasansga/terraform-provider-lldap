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

# Random password for test user
resource "random_password" "user_password" {
  length  = 16
  special = true
}

# Create a test user
resource "lldap_user" "test_user" {
  username     = "testuser-${random_string.suffix.result}"
  email        = "testuser-${random_string.suffix.result}@example.com"
  display_name = "Test User ${random_string.suffix.result}"
  first_name   = "Test"
  last_name    = "User"
  password     = random_password.user_password.result
}

# Create a custom user attribute
resource "lldap_user_attribute" "test_attr" {
  name           = "testuserattr${random_string.suffix.result}"
  attribute_type = var.test_attribute_type
  is_list        = var.test_is_list
  is_visible     = true
  is_editable    = true
}

# Assign values to the custom attribute for the user
resource "lldap_user_attribute_assignment" "test_assignment" {
  user_id      = lldap_user.test_user.id
  attribute_id = lldap_user_attribute.test_attr.id
  value        = var.test_attribute_type == "JPEG_PHOTO" ? [filebase64("${path.module}/test.jpeg")] : var.test_values
}

# Data sources for validation
data "lldap_user" "test_user" {
  id = lldap_user.test_user.id
  depends_on = [lldap_user_attribute_assignment.test_assignment]
}

# Outputs for testing
output "test_user" {
  value = lldap_user.test_user
  sensitive = true
}

output "test_user_attr" {
  value = lldap_user_attribute.test_attr
}

output "test_assignment" {
  value = lldap_user_attribute_assignment.test_assignment
}

output "user_with_attributes" {
  value = data.lldap_user.test_user
  sensitive = true
}
