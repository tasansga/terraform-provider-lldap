terraform {
  required_providers {
    lldap = {
      source  = "tasansga/lldap"
      version = "0.0.1"
    }
  }
}

variable "lldap_base_dn" {}

provider "lldap" {
  base_dn  = var.lldap_base_dn
}

data "lldap_users" "users" {}

output "users" {
  value = data.lldap_users.users.users
}
