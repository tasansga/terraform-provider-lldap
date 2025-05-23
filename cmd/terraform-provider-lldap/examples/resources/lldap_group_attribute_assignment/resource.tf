resource "lldap_group" "test" {
  display_name = "Test group"
}

resource "lldap_group_attribute" "test" {
  name           = "test-change"
  attribute_type = "STRING"
  is_list        = true
  is_visible     = false
}

resource "lldap_group_attribute_assignment" "test" {
  group_id     = lldap_group.group.id
  attribute_id = lldap_group_attribute.test.id
  value        = ["attribute value for group ${lldap_group.test.display_name} and attribute ${lldap_group_attribute.test.name}"]
}
