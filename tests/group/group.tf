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

resource "lldap_group" "test" {
  display_name = "Test group"
}

output "test" {
  value = lldap_group.test
}
