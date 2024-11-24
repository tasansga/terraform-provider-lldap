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

resource "lldap_group" "group" {
  display_name = "Test group"
}

resource "lldap_group_attribute" "test" {
  count          = 50
  name           = "test-${count.index}"
  attribute_type = "STRING"
}

output "group_attr" {
  value = lldap_group_attribute.test
}

resource "lldap_group_attribute" "test_change" {
  name           = "test-change"
  attribute_type = "STRING"
  is_list        = true
  is_visible     = false
}

output "group_attr_change" {
  value = lldap_group_attribute.test_change
}

resource "lldap_group_attribute_assignment" "test" {
  group_id     = lldap_group.group.id
  attribute_id = lldap_group_attribute.test_change.id
  value        = ["test-value: ${lldap_group.group.display_name}"]
}

output "group_attr_assignment" {
  value = lldap_group_attribute_assignment.test
}
