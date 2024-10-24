terraform {
  required_providers {
    lldap = {
      source  = "tasansga/lldap"
      version = "0.0.1"
    }
  }
}

variable "lldap_url" {}
variable "lldap_username" {}
variable "lldap_password" {}

provider "lldap" {
  lldap_url      = var.lldap_url
  lldap_username = var.lldap_username
  lldap_password = var.lldap_password
}

resource "lldap_user" "user1" {
  username     = "user1"
  display_name = "User 1"
  email        = "user1@this.test"
  first_name   = "FIRST"
  last_name    = "LAST"
}

output "user1" {
  value = lldap_user.user1
}

resource "lldap_user" "user2" {
  username     = "user2"
  display_name = "User 2"
  email        = "user2@this.test"
}

data "lldap_users" "users" {
  depends_on = [lldap_user.user1, lldap_user.user2]
}

output "user2" {
  value = lldap_user.user2
}

output "users_map" {
  value = { for user in data.lldap_users.users.users : user.id => user }
}
