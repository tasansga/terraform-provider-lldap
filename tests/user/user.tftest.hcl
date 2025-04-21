run "test_user" {
  assert {
    condition     = output.user[0].id == "user0"
    error_message = "id should be 'user0'"
  }
  assert {
    condition     = output.user[0].username == "user0"
    error_message = "username should be 'user0'"
  }
  assert {
    condition     = output.user_password[0] == output.user[0].password
    error_message = "password mapping failed"
  }
  assert {
    condition     = output.user[0].email == "user0@this.test"
    error_message = "email should be 'user0@this.test'"
  }
  assert {
    condition     = output.user[0].avatar == output.avatar_base64
    error_message = "invalid value for avatar base64"
  }
  assert {
    condition     = output.user[0].display_name == "User 0"
    error_message = "display_name should be 'User 0'"
  }
  assert {
    condition     = output.user[0].first_name == "FIRST 0"
    error_message = "first_name should be 'FIRST 0'"
  }
  assert {
    condition     = output.user[0].last_name == "LAST 0"
    error_message = "last_name should be 'LAST 0'"
  }
  assert {
    condition     = output.user[0].creation_date != null && output.user[0].creation_date != ""
    error_message = "creation_date should not be null or empty string"
  }
  assert {
    condition     = output.user[0].uuid != null && output.user[0].uuid != ""
    error_message = "uuid should not be null or empty string"
  }
  assert {
    condition = toset(keys({ for k, v in output.user[0].attributes : v.name => v.value })) == toset(["avatar","creation_date","display_name","first_name","last_name","mail","user_id","uuid"])
    error_message = "Error matching attributes: ${nonsensitive(jsonencode(keys({ for k, v in output.user[0].attributes : v.name => v.value })))}"
  }
}

run "test_users_map" {
  // Map is a data source so we need to retest the values from the resource
  assert {
    condition     = length(keys(output.users_map)) == 1 + local.user_count + 1 // default "admin" plus our count new users plus newpasswd user
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
    condition     = output.users_map["user1"].username == "user1"
    error_message = "username should be 'user1'"
  }
  assert {
    condition     = output.users_map["user1"].first_name == "FIRST 1"
    error_message = "user1 first_name should be 'FIRST 1'"
  }
  assert {
    condition     = output.users_map["user1"].last_name == "LAST 1"
    error_message = "user1 last_name should be 'LAST 1'"
  }
  assert {
    condition     = output.users_map["user1"].creation_date != null && output.users_map["user1"].creation_date != ""
    error_message = "user1 creation_date should not be null or empty string"
  }
  assert {
    condition     = output.users_map["user1"].uuid != null && output.users_map["user1"].uuid != ""
    error_message = "user1 uuid should not be null or empty string"
  }
  assert {
    condition     = output.users_map["user1"].avatar == output.avatar_base64
    error_message = "Invalid value for user avatar base64"
  }
}

run "test_nopasswd_user" {
  assert {
    condition     = output.nopasswd.password == null
    error_message = "invalid value for no password"
  }
}
