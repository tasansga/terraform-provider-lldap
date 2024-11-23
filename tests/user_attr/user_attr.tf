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

resource "random_password" "user" {
  length = 16
}

resource "lldap_user" "user" {
  username     = "user_attr"
  email        = "user_attr@this.test"
  password     = random_password.user.result
  display_name = "User Attr"
  first_name   = "FIRST ATTR"
  last_name    = "LAST ATTR"
}

resource "lldap_user_attribute" "test" {
  count          = 50
  name           = "test-change-${count.index}"
  attribute_type = "STRING"
  is_editable    = true
  is_list        = true
  is_visible     = false
}

output "user_attr" {
  value = lldap_user_attribute.test
}

resource "lldap_user_attribute_assignment" "test" {
  for_each     = toset([ for attr in lldap_user_attribute.test : attr.name ])
  user_id      = lldap_user.user.id
  attribute_id = each.value
  value        = ["test-value: ${each.value}"]
}

output "user_attr_assignment" {
  value = lldap_user_attribute_assignment.test
}
