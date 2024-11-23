run "test_user_attr" {
  assert {
    condition     = length(toset(output.user_attr)) == 50
    error_message = "unexpected length of result"
  }
}

run "test_user_attr_assignment" {
  assert {
    condition     = length(output.user_attr_assignment) == 50
    error_message = "wrong number of user attribute assignments"
  }
  assert {
    condition = jsondecode(jsonencode(output.user_attr_assignment["test-change-0"])) == jsondecode(jsonencode({
      "attribute_id" : "test-change-0",
      "id" : "user_attr:test-change-0",
      "user_id" : "user_attr",
      "value" : ["test-value: test-change-0"]
    }))
    error_message = "unexpected value for user attribute assignment: ${jsonencode(output.user_attr_assignment["test-change-0"])}"
  }
}
