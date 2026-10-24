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

# Random suffix for unique naming
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

variable "password" {
  type     = string
  nullable = true
  default = null
}

variable "password_wo" {
  type     = string
  nullable = true
  default = null
}

variable "password_wo_version" {
  type     = string
  nullable = true
  default = null
}

resource "lldap_user" "this" {
  username     = "user-${random_string.suffix.result}"
  email        = "user-${random_string.suffix.result}@this.test"
  display_name = "User ${random_string.suffix.result}"

  password            = var.password
  password_wo         = var.password_wo
  password_wo_version = var.password_wo_version
}

output "username" {
  value = lldap_user.this.username
}
