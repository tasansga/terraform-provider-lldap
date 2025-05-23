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

resource "random_password" "user_password" {
  length = 16
}

# Variable for controlling member count
variable "member_count" {
  type    = number
  default = 3
  validation {
    condition     = var.member_count >= 0 && var.member_count <= 10
    error_message = "member_count must be between 0 and 10."
  }
}

# Create multiple users for membership testing
resource "random_password" "user_passwords" {
  count  = var.member_count
  length = 16
}

resource "lldap_user" "test_users" {
  count        = var.member_count
  username     = "testuser${count.index}-${random_string.test.result}"
  email        = "testuser${count.index}-${random_string.test.result}@member.test"
  password     = random_password.user_passwords[count.index].result
  display_name = "Test User ${count.index}"
  first_name   = "Test"
  last_name    = "User${count.index}"
}

resource "lldap_user" "user" {
  username     = "user-${random_string.test.result}"
  email        = "user-${random_string.test.result}@this.test"
  password     = random_password.user_password.result
  display_name = "User ${random_string.test.result}"
  first_name   = "FIRST"
  last_name    = "LAST"
}

resource "lldap_group" "group" {
  display_name = "Test group ${random_string.test.result}"
}

# Create multiple groups for membership testing
resource "lldap_group" "test_groups" {
  count        = var.member_count
  display_name = "Test Group ${count.index} - ${random_string.test.result}"
}

resource "lldap_member" "member" {
  group_id = lldap_group.group.id
  user_id  = lldap_user.user.id
}

# Create multiple member relationships
resource "lldap_member" "test_members" {
  count    = var.member_count
  group_id = lldap_group.test_groups[count.index].id
  user_id  = lldap_user.test_users[count.index].id
}

# Data sources for verification
data "lldap_group" "test_group" {
  id = lldap_group.group.id
  depends_on = [lldap_member.member]
}

data "lldap_user" "test_user" {
  id = lldap_user.user.id
  depends_on = [lldap_member.member]
}

data "lldap_groups" "all_groups" {
  depends_on = [lldap_group.group, lldap_group.test_groups]
}

data "lldap_users" "all_users" {
  depends_on = [lldap_user.user, lldap_user.test_users]
}

# Individual data sources for detailed verification
data "lldap_group" "test_groups_detailed" {
  count = var.member_count
  id = lldap_group.test_groups[count.index].id
  depends_on = [lldap_member.test_members]
}

data "lldap_user" "test_users_detailed" {
  count = var.member_count
  id = lldap_user.test_users[count.index].id
  depends_on = [lldap_member.test_members]
}

# Outputs for verification
output "user_id" {
  value = lldap_user.user.id
}

output "group_id" {
  value = lldap_group.group.id
}

output "member" {
  value = lldap_member.member
}

output "test_user" {
  value     = lldap_user.user
  sensitive = true
}

output "test_user_detailed" {
  value     = data.lldap_user.test_user
  sensitive = true
}

output "test_group" {
  value = lldap_group.group
}

output "test_group_detailed" {
  value = data.lldap_group.test_group
}

output "all_users" {
  value = data.lldap_users.all_users.users
}

output "all_groups" {
  value = data.lldap_groups.all_groups.groups
}

output "test_users" {
  value     = lldap_user.test_users
  sensitive = true
}

output "test_users_detailed" {
  value     = data.lldap_user.test_users_detailed
  sensitive = true
}

output "test_groups" {
  value = lldap_group.test_groups
}

output "test_groups_detailed" {
  value = data.lldap_group.test_groups_detailed
}

output "test_members" {
  value = lldap_member.test_members
}

output "member_count" {
  value = var.member_count
}
