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

variable "test_is_visible" {
  description = "Whether the attribute is visible"
  type        = bool
  default     = true
}

# Test group attribute resource
resource "lldap_group_attribute" "test_attr" {
  name           = "testgroupattr${random_string.suffix.result}"
  attribute_type = var.test_attribute_type
  is_list        = var.test_is_list
  is_visible     = var.test_is_visible
}

# Data sources for validation
data "lldap_group_attributes" "group_attrs" {
  depends_on = [lldap_group_attribute.test_attr]
}

# Outputs for testing
output "test_group_attr" {
  value = lldap_group_attribute.test_attr
}

output "group_attrs" {
  value = data.lldap_group_attributes.group_attrs
}
