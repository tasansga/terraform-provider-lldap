terraform {
  required_providers {
    lldap = {
      source  = "tasansga/lldap"
      version = "0.0.1"
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

locals {
  group_count = 20
}

resource "lldap_group" "test" {
  count        = local.group_count
  display_name = "Test group ${count.index}"
}

data "lldap_groups" "test" {
  depends_on = [lldap_group.test]
}

output "group_count" {
  value = local.group_count
}

output "data" {
  value = data.lldap_groups.test
}

output "test" {
  value = lldap_group.test
}
