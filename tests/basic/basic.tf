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

data "lldap_users" "users" {}

output "users" {
  value = data.lldap_users.users.users
}

output "users_map" {
  value = { for user in data.lldap_users.users.users : user.id => user }
}

data "lldap_groups" "groups" {}

output "groups" {
  value = data.lldap_groups.groups
}

output "groups_map" {
  value = { for group in data.lldap_groups.groups.groups : group.display_name => group }
}

data "lldap_group" "lldap_admin" {
  id = 1
}

output "group" {
  value = data.lldap_group.lldap_admin
}

data "lldap_user" "admin" {
  id = "admin"
}

output "user" {
  value = data.lldap_user.admin
}


data "lldap_user_attributes" "user_attrs" {}

output "user_attrs" {
  value = data.lldap_user_attributes.user_attrs
}
