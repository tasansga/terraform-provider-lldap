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

resource "random_string" "random" {
  length  = 8
  special = false
  upper   = false
}

resource "lldap_group" "group" {
  display_name = "Test group ${random_string.random.result}"
}

resource "random_password" "user" {
  length = 16
}

resource "lldap_user" "user" {
  count        = 10
  username     = "user-${random_string.random.result}-${count.index}"
  email        = "user-${random_string.random.result}-${count.index}@this.test"
  password     = random_password.user.result
  display_name = "User"
  first_name   = "FIRST"
  last_name    = "LAST"
}

resource "lldap_group_memberships" "group" {
  group_id = lldap_group.group.id
  user_ids = toset([ for user in lldap_user.user: user.id ])
}

output "group_memberships" {
  value = lldap_group_memberships.group
}
