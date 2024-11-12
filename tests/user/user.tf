terraform {
  required_providers {
    lldap = {
      source  = "tasansga/lldap"
      version = "0.0.1"
    }
    random = {
      source = "hashicorp/random"
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

locals {
  user_count = 20
}

resource "random_password" "user" {
  count  = local.user_count
  length = 16
}

output "user_count" {
  value = local.user_count
}

resource "lldap_user" "user" {
  for_each     = { for k, v in random_password.user : k => v.result }
  username     = "user${each.key}"
  email        = "user${each.key}@this.test"
  password     = each.value
  display_name = "User ${each.key}"
  first_name   = "FIRST ${each.key}"
  last_name    = "LAST ${each.key}"
  avatar       = filebase64("${path.module}/test.jpeg")
}

output "user" {
  value     = lldap_user.user
  sensitive = true
}

output "user_password" {
  value     = [for k, v in random_password.user : v.result]
  sensitive = true
}

output "avatar_base64" {
  value = filebase64("${path.module}/test.jpeg")
}

data "lldap_users" "users" {
  depends_on = [lldap_user.user]
}

output "users_map" {
  value = { for user in data.lldap_users.users.users : user.id => user }
}
