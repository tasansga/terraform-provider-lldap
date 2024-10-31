run "test_group" {
  assert {
    condition     = output.test[0].id != null && output.test[0].id != 0
    error_message = "id should not be null or zero"
  }
  assert {
    condition     = output.test[1].display_name == "Test group 1"
    error_message = "display_name should be 'Test group'"
  }
  assert {
    condition     = output.test[1].creation_date != null && output.test[1].creation_date != ""
    error_message = "creation_date should not be null or empty string"
  }
  assert {
    condition     = output.test[1].uuid != null && output.test[1].uuid != ""
    error_message = "uuid should not be null or empty string"
  }
  assert {
    condition     = output.test[0].users != null
    error_message = "users should not be null"
  }
  assert {
    condition = length(data.lldap_groups.test.groups) == output.group_count + 3 // new groups plus lldap_admin, lldap_password_manager, lldap_strict_readonly
    error_message = "Found ${length(data.lldap_groups.test.groups)} groups (${jsonencode(data.lldap_groups.test)}) in data source but should be ${output.group_count + 3}"
  }
}
