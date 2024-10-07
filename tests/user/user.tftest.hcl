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