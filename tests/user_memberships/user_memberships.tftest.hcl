run "test_user_memberships" {
  assert {
    condition     = length(output.user_memberships.group_ids) == 10
    error_message = "group list should contain exactly 10 groups"
  }
}
