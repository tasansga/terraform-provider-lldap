resource "lldap_user" "test" {
  display_name = "Test user"
}

resource "lldap_user_attribute" "test" {
  name           = "test-change"
  attribute_type = "STRING"
  is_list        = true
  is_visible     = false
}

resource "lldap_user_attribute_assignment" "test" {
  user_id      = lldap_user.user.id
  attribute_id = lldap_user_attribute.test.id
  value        = ["attribute value for user ${lldap_user.test.display_name} and attribute ${lldap_user_attribute.test.name}"]
}
