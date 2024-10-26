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
  assert {
    condition     = output.user1.uuid != null && output.user1.uuid != ""
    error_message = "uuid should not be null or empty string"
  }
}

run "test_users_map" {
  assert {
    condition     = length(keys(output.users_map)) == 3 // default "admin" plus our new two users"
    error_message = "could not find users"
  }
  assert {
    condition     = output.users_map["user1"].display_name == "User 1"
    error_message = "user1 display name error"
  }
  assert {
    condition     = output.users_map["user1"].email == "user1@this.test"
    error_message = "user1 email error"
  }
  assert {
    condition     = output.users_map["user2"].display_name == "User 2"
    error_message = "user2 display name error"
  }
  assert {
    condition     = output.users_map["user2"]["email"] == "user2@this.test"
    error_message = "user2 email error"
  }
  assert {
    condition     = output.users_map["user1"].avatar == output.avatar_base64
    error_message = "Invalid value for user avatar base64"
  }
}
