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

resource "random_string" "email_prefix" {
  length  = 8
  special = false
  upper   = false
}

resource "lldap_user" "user" {
  username     = "user"
  email        = "user-${random_string.email_prefix.result}@this.test"
  password     = random_password.user.result
  display_name = "User"
  first_name   = "FIRST"
  last_name    = "LAST"
}

variable "nopasswd" {
  type    = string
  default = null
}

resource "lldap_user" "nopasswd" {
  username     = "user-nopasswd-change"
  email        = "nopasswd-${random_string.email_prefix.result}@this.test"
  password     = var.nopasswd
}
