run "test_user" {
  assert {
    condition     = output.user1.id == "user1"
    error_message = "id should be 'user1'"
  }
  assert {
    condition     = output.user1.username == "user1"
    error_message = "username should be 'user1'"
  }
  assert {
    condition     = output.user1.email == "user1@this.test"
    error_message = "email should be 'user1@this.test'"
  }
  assert {
    condition     = output.user1.display_name == "User 1"
    error_message = "display_name should be 'User 1'"
  }
  assert {
    condition     = output.user1.first_name == "FIRST"
    error_message = "first_name should be 'FIRST'"
  }
  assert {
    condition     = output.user1.last_name == "LAST"
    error_message = "last_name should be 'LAST'"
  }
  assert {
    condition     = output.user1.creation_date != null && output.user1.creation_date != ""
    error_message = "creation_date should not be null or empty string"
  }
}

/*
run "test_users_map" {
  assert {
    condition     = length(output.users) == 4
    error_message = "Did not find expected users"
  }
  assert {
    condition     = output.users_map["user1"]["display_name"] == "User 1"
    error_message = "user1 display name error"
  }
  assert {
    condition     = output.users_map["user1"]["email"] == "user1@this.test"
    error_message = "user1 email error"
  }
  assert {
    condition     = output.users_map["user2"]["display_name"] == "User 2"
    error_message = "user2 display name error"
  }
  assert {
    condition     = output.users_map["user2"]["email"] == "user2@this.test"
    error_message = "user2 email error"
  }
  assert {
    condition     = output.users_map["user3"]["display_name"] == "User 3"
    error_message = "user3 display name error"
  }
  assert {
    condition     = output.users_map["user3"]["email"] == "user3@this.test"
    error_message = "user3 email error"
  }
}

run "test_users" {
    assert {
        condition = output.user1 == output.users_map["user1"]
        error_message = "output mismatch user1"
    }
    assert {
        condition = output.user2 == output.users_map["user2"]
        error_message = "output mismatch user2"
    }
    assert {
        condition = output.user3 == output.users_map["user3"]
        error_message = "output mismatch user3"
    }
}
*/
