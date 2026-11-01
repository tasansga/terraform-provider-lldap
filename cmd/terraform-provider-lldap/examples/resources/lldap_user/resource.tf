# Manage an user with attributes
resource "lldap_user" "user" {
  username     = "myuser"
  email        = "myuser@in.the.test"
  password     = "super-secret password!"
  display_name = "My User"
  first_name   = "My"
  last_name    = "User"
  avatar       = filebase64("${path.module}/myuser.jpeg")
}

# Example using write-only password (password not stored in state)
resource "lldap_user" "user" {
  username            = "myuser"
  email               = "myuser@in.the.test"
  display_name        = "My User"
  password_wo         = var.password # typically sourced from a secret manager or similar
  password_wo_version = "1"
}
