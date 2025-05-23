# Get the default admin user
data "lldap_user" "admin" {
  id = "admin"
}