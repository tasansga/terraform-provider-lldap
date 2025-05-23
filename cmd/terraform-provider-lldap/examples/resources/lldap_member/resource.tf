# Get the default admin user
data "lldap_user" "admin" {
  id = "admin"
}

# Create and manage a new group
resource "lldap_group" "test_admin" {
  display_name = "Test admin group"
}

# Manage the membership for the user in the group
resource "lldap_member" "test_admin" {
  group_id = lldap_group.test_admin.id
  user_id  = data.lldap_user.admin.id
}
