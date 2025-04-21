resource "lldap_user" "nopasswd" {
  username     = "user-nopasswd"
  email        = "user-nopasswd@this.test"
}

output "nopasswd" {
  value     = lldap_user.nopasswd
  sensitive = true
}
