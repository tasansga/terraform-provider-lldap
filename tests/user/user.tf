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

resource "random_password" "user" {
  length = 16
}

resource "random_string" "email_prefix" {
  length  = 8
  special = false
  upper   = false
}

# Variable for controlling user count
variable "user_count" {
  type    = number
  default = 3
  validation {
    condition     = var.user_count >= 0 && var.user_count <= 10
    error_message = "user_count must be between 0 and 10."
  }
}

# Create multiple users for lifecycle testing
resource "random_password" "user_passwords" {
  count  = var.user_count
  length = 16
}

resource "lldap_user" "test_users" {
  count        = var.user_count
  username     = "testuser${count.index}-${random_string.suffix.result}"
  email        = "testuser${count.index}-${random_string.suffix.result}@user.test"
  password     = random_password.user_passwords[count.index].result
  display_name = "Test User ${count.index}"
  first_name   = "Test"
  last_name    = "User${count.index}"
}

# Main test user with comprehensive attributes
resource "lldap_user" "user" {
  username     = "user-${random_string.suffix.result}"
  email        = "user-${random_string.email_prefix.result}@this.test"
  password     = random_password.user.result
  display_name = "User ${random_string.suffix.result}"
  avatar       = filebase64("${path.module}/test.jpeg")
  first_name   = "FIRST"
  last_name    = "LAST"
}

variable "nopasswd" {
  type    = string
  default = null
}

resource "lldap_user" "nopasswd" {
  username     = "user-nopasswd-${random_string.suffix.result}"
  email        = "nopasswd-${random_string.email_prefix.result}@this.test"
  password     = var.nopasswd
  display_name = "No Password User"
  first_name   = "No"
  last_name    = "Password"
}

# Test user with custom attributes
variable "create_user_with_attrs" {
  type    = bool
  default = true
}

resource "lldap_user_attribute" "test_attr" {
  count          = var.create_user_with_attrs ? 1 : 0
  name           = "test-user-attr-${random_string.suffix.result}"
  attribute_type = "STRING"
  is_list        = false
  is_visible     = true
  is_editable    = true
}

resource "lldap_user" "user_with_attrs" {
  count        = var.create_user_with_attrs ? 1 : 0
  username     = "userattrs-${random_string.suffix.result}"
  email        = "userattrs-${random_string.suffix.result}@attrs.test"
  password     = random_password.user.result
  display_name = "User with Attributes"
  first_name   = "Attrs"
  last_name    = "User"

  depends_on = [lldap_user_attribute.test_attr]
}

# Data sources for verification
data "lldap_users" "all_users" {
  depends_on = [lldap_user.test_users, lldap_user.user, lldap_user.nopasswd, lldap_user.user_with_attrs]
}

data "lldap_user_attributes" "all_user_attrs" {
  depends_on = [lldap_user_attribute.test_attr]
}

# Individual user data source for detailed verification
data "lldap_user" "main_user" {
  id = lldap_user.user.id
  depends_on = [lldap_user.user]
}

# Outputs for verification
output "user_count" {
  value = var.user_count
}

output "created_users" {
  value     = lldap_user.test_users
  sensitive = true
}

output "all_users" {
  value = data.lldap_users.all_users.users
}

output "main_user" {
  value     = lldap_user.user
  sensitive = true
}

output "main_user_detailed" {
  value     = data.lldap_user.main_user
  sensitive = true
}

output "nopasswd_user" {
  value     = lldap_user.nopasswd
  sensitive = true
}

output "user_with_attrs" {
  value     = var.create_user_with_attrs ? lldap_user.user_with_attrs[0] : null
  sensitive = true
}

output "test_attr" {
  value = var.create_user_with_attrs ? lldap_user_attribute.test_attr[0] : null
}

output "all_user_attrs" {
  value = data.lldap_user_attributes.all_user_attrs
}
