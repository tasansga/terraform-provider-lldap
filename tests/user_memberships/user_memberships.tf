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

variable "num_groups" {
  default = 10
}

variable "max_groups" {
  default = 10
}

provider "lldap" {
  http_url = var.lldap_http_url
  ldap_url = var.lldap_ldap_url
  username = var.lldap_username
  password = var.lldap_password
  base_dn  = var.lldap_base_dn
}

resource "random_password" "user" {
  length = 16
}

resource "random_string" "random" {
  length  = 8
  special = false
  upper   = false
}

resource "lldap_user" "user" {
  username     = "user-${random_string.random.result}"
  email        = "user-${random_string.random.result}@this.test"
  password     = random_password.user.result
  display_name = "User"
  first_name   = "FIRST"
  last_name    = "LAST"
}

resource "lldap_group" "test" {
  count        = var.max_groups
  display_name = "Test group ${count.index}"
}

resource "lldap_user_memberships" "user" {
  user_id    = lldap_user.user.id
  group_ids  = toset([ for group in slice(lldap_group.test,0, var.num_groups): group.id ])
  depends_on = [lldap_user.user]
}

# Data sources for verification
data "lldap_user" "test_user" {
  id = lldap_user.user.id
  depends_on = [lldap_user_memberships.user]
}

data "lldap_users" "all_users" {
  depends_on = [lldap_user.user]
}

data "lldap_groups" "all_groups" {
  depends_on = [lldap_group.test]
}

# Individual group data sources to verify memberships
data "lldap_group" "test_groups" {
  count = var.num_groups
  id = lldap_group.test[count.index].id
  depends_on = [lldap_user_memberships.user]
}

# Outputs for verification
output "user_memberships" {
  value = lldap_user_memberships.user
}

output "test_user" {
  value     = lldap_user.user
  sensitive = true
}

output "test_user_detailed" {
  value     = data.lldap_user.test_user
  sensitive = true
}

output "all_users" {
  value = data.lldap_users.all_users.users
}

output "all_groups" {
  value = data.lldap_groups.all_groups.groups
}

output "test_groups" {
  value = lldap_group.test
}

output "test_groups_detailed" {
  value = data.lldap_group.test_groups
}

output "num_groups" {
  value = var.num_groups
}

output "max_groups" {
  value = var.max_groups
}
