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