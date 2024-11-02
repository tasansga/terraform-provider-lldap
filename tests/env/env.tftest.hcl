run "test_users" {
  assert {
    condition     = length(output.users) == 1
    error_message = "Did not find expected user"
  }
  assert {
    condition     = tolist(output.users)[0].display_name == "Administrator"
    error_message = "Could not find admin user"
  }
}
