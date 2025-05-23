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

variable "num_users" {
  default = 10
}

variable "max_users" {
  default = 10
}

variable "enable_out_of_band_user" {
  default = false
}

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
  count        = var.max_users
  username     = "user-${random_string.random.result}-${count.index}"
  email        = "user-${random_string.random.result}-${count.index}@this.test"
  password     = random_password.user.result
  display_name = "User"
  first_name   = "FIRST"
  last_name    = "LAST"
}

resource "lldap_user" "out_of_band" {
  username     = "user-${random_string.random.result}-oob"
  email        = "user-${random_string.random.result}-oob@this.test"
  password     = random_password.user.result
  display_name = "User"
  first_name   = "FIRST"
  last_name    = "LAST"
}

resource "lldap_member" "out_of_band" {
  count    = var.enable_out_of_band_user ? 1 : 0
  group_id = lldap_group.group.id
  user_id  = lldap_user.out_of_band.id
}

resource "lldap_group_memberships" "group" {
  group_id = lldap_group.group.id
  user_ids = toset([ for user in slice(lldap_user.user,0, var.num_users): user.id ])
}

# Data sources for verification
data "lldap_group" "test_group" {
  id = lldap_group.group.id
  depends_on = [lldap_group_memberships.group, lldap_member.out_of_band]
}

data "lldap_groups" "all_groups" {
  depends_on = [lldap_group.group]
}

data "lldap_users" "all_users" {
  depends_on = [lldap_user.user, lldap_user.out_of_band]
}

# Individual user data sources to verify memberships
data "lldap_user" "test_users" {
  count = var.num_users
  id = lldap_user.user[count.index].id
  depends_on = [lldap_group_memberships.group]
}

# Outputs for verification
output "group_memberships" {
  value = lldap_group_memberships.group
}

output "test_group" {
  value = lldap_group.group
}

output "test_group_detailed" {
  value = data.lldap_group.test_group
}

output "all_groups" {
  value = data.lldap_groups.all_groups.groups
}

output "all_users" {
  value = data.lldap_users.all_users.users
}

output "test_users" {
  value     = lldap_user.user
  sensitive = true
}

output "test_users_detailed" {
  value     = data.lldap_user.test_users
  sensitive = true
}

output "out_of_band_user" {
  value     = lldap_user.out_of_band
  sensitive = true
}

output "out_of_band_member" {
  value = var.enable_out_of_band_user ? lldap_member.out_of_band[0] : null
}

output "num_users" {
  value = var.num_users
}

output "max_users" {
  value = var.max_users
}
