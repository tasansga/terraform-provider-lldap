run "test_member" {
  assert {
    condition = jsonencode(output.test) == jsonencode({
      "id" : "4:admin",
      "group_id" : 4,
      "user_id" : "admin"
      "group_display_name" : "Test member group",
    })
    error_message = "Member check failed"
  }
  assert {
    condition     = tolist(output.user.groups)[0].display_name == "Test member group"
    error_message = "no or wrong computed group membership: ${nonsensitive(jsonencode(output.user))}"
  }
}
