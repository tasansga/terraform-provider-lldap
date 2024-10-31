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

data "lldap_user" "admin" {
  id = "admin"
}

resource "lldap_group" "test_member" {
  display_name = "Test member group"
}

resource "lldap_member" "test" {
  group_id = lldap_group.test_member.id
  user_id  = data.lldap_user.admin.id
}

output "test" {
  value = lldap_member.test
}
