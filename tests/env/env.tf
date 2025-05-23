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

variable "lldap_base_dn" {}



# Test basic data source with env vars
data "lldap_users" "users" {}

# Create test resources to verify env var authentication works
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

resource "random_password" "password" {
  length = 16
}

# Test user creation with environment variable authentication
variable "create_user" {
  type    = bool
  default = true
}

resource "lldap_user" "env_test_user" {
  count        = var.create_user ? 1 : 0
  username     = "envtest-${random_string.suffix.result}"
  email        = "envtest-${random_string.suffix.result}@env.test"
  password     = random_password.password.result
  display_name = "Environment Test User"
  first_name   = "Env"
  last_name    = "Test"
}

# Test group creation with environment variable authentication
variable "create_group" {
  type    = bool
  default = true
}

resource "lldap_group" "env_test_group" {
  count        = var.create_group ? 1 : 0
  display_name = "Environment Test Group ${random_string.suffix.result}"
}



# Standard provider configuration - uses variables from test.auto.tfvars
provider "lldap" {
  http_url = var.lldap_http_url
  ldap_url = var.lldap_ldap_url
  username = var.lldap_username
  password = var.lldap_password
  base_dn  = var.lldap_base_dn
}

# Variables for explicit authentication testing
variable "lldap_http_url" {
  type    = string
  default = ""
}

variable "lldap_ldap_url" {
  type    = string
  default = ""
}

variable "lldap_username" {
  type    = string
  default = ""
}

variable "lldap_password" {
  type    = string
  default = ""
  sensitive = true
}

# Outputs for verification
output "users" {
  value = data.lldap_users.users.users
}

output "user_count" {
  value = length(data.lldap_users.users.users)
}

output "env_test_user" {
  value     = var.create_user ? lldap_user.env_test_user[0] : null
  sensitive = true
}

output "env_test_group" {
  value = var.create_group ? lldap_group.env_test_group[0] : null
}

output "test_method" {
  value = "environment-variables"
}

output "base_dn" {
  value = var.lldap_base_dn
}
