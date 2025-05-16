run "test_group_memberships" {
  assert {
    condition     = length(output.group_memberships.user_ids) == 10
    error_message = "memberships list should contain exactly 10 users"
  }
}
