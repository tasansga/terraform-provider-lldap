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
