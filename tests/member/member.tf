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

resource "random_password" "user" {
  length = 16
}

resource "lldap_user" "user" {
  username     = "user"
  email        = "user@this.test"
  password     = random_password.user.result
  display_name = "User"
  first_name   = "FIRST"
  last_name    = "LAST"
}

resource "lldap_group" "test_member" {
  display_name = "Test member group"
}

resource "lldap_member" "test" {
  group_id = lldap_group.test_member.id
  user_id  = data.lldap_user.admin.id
}

resource "lldap_member" "user" {
  group_id = lldap_group.test_member.id
  user_id  = lldap_user.user.id
}

data "lldap_user" "user" {
  // We need this data source because the membership is created
  // AFTER the user has been created, i.e. the user's group list
  // in the resource doesn't contain the membership yet.
  id         = lldap_user.user.id
  depends_on = [lldap_member.user]
}

output "test" {
  value = lldap_member.test
}

output "user" {
  value     = data.lldap_user.user
  sensitive = true

}
