run "test_group_attr" {
  assert {
    condition     = length(output.group_attr) == 50
    error_message = "unexpected length of result"
  }
  assert {
    condition = alltrue([
      for attr in output.group_attr : attr.is_list == false && attr.is_visible == true
    ])
    error_message = jsonencode(output.group_attr)
  }
}

run "test_group_attr_change" {
  assert {
    condition     = output.group_attr_change.is_list == true && output.group_attr_change.is_visible == false
    error_message = jsonencode(output.group_attr)
  }
}

run "test_group_attr_assignment" {
  assert {
    condition = jsondecode(jsonencode(output.group_attr_assignment)) == jsondecode(jsonencode({
      "attribute_id" : "test-change",
      "group_id" : 4,
      "id" : "4:test-change",
      "value" : ["test-value: Test group"]
    }))
    error_message = "unexpected value for user attribute assignment: ${jsonencode(output.group_attr_assignment)}"
  }
}
