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

# Test user attribute resource with camelCase name
resource "lldap_user_attribute" "ssh_public_key" {
  name           = "sshPublicKey"
  attribute_type = "STRING"
  is_list        = true
  is_visible     = true
  is_editable    = false
}

# Data source for validation
data "lldap_user_attributes" "user_attrs" {
  depends_on = [lldap_user_attribute.ssh_public_key]
}

# Outputs for testing
output "test_user_attr" {
  value = lldap_user_attribute.ssh_public_key
}

output "user_attrs" {
  value = data.lldap_user_attributes.user_attrs
}
