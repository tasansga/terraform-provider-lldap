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

resource "random_string" "test" {
  length  = 8
  special = false
  upper   = false
}

resource "lldap_user" "user" {
  username     = "user"
  email        = "user-${random_string.test.result}@this.test"
  display_name = "User"
  first_name   = "FIRST"
  last_name    = "LAST"
}

output "user_id" {
  value = lldap_user.user.id
}

resource "lldap_group" "group" {
  display_name = "Test group ${random_string.test.result}"
}

output "group_id" {
  value = lldap_group.group.id
}

resource "lldap_member" "member" {
  group_id = lldap_group.group.id
  user_id  = lldap_user.user.id
}
