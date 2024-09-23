run "test_users" {
  assert {
    condition     = length(output.users) >= 1
    error_message = "Did not find expected users"
  }
  assert {
    condition     = output.users_map["admin"]["display_name"] == "Administrator"
    error_message = "Could not find admin user"
  }
}

run "test_groups" {
  assert {
    condition     = length(output.groups) >= 1
    error_message = "Did not find expected groups"
  }
  assert {
    condition     = alltrue([
        for group in ["lldap_admin", "lldap_password_manager", "lldap_strict_readonly"]:
        contains(keys(output.groups_map), group)
    ])
    error_message = "Could not find required groups"
  }
}

run "test_group" {
    assert {
        condition = output.group.display_name == "lldap_admin"
        error_message = "Mapping of group display name failed"
    }
    assert {
        condition = length(output.group.users) == 1
        error_message = "Mapping of group users failed"
    }
    assert {
        condition = { for user in output.group.users: user.id => user }["admin"].display_name == "Administrator"
        error_message = "Mapping of group user details failed"
    }
}

run "test_user" {
    assert {
        condition = output.user.display_name == "Administrator"
        error_message = "Mapping of user display name failed"
    }
    assert {
        condition = length(output.user.groups) == 1
        error_message = "Mapping of user groups failed"
    }
    assert {
        condition = { for group in output.user.groups: "${group.id}" => group.display_name }["1"] == "lldap_admin"
        error_message = "Mapping of user groups details failed"
    }
}
