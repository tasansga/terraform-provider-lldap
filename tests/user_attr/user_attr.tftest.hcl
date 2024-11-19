run "test_user_attr" {
  assert {
    condition     = length(output.user_attr) == 50
    error_message = "unexpected length of result"
  }
  assert {
    condition = alltrue([
      for attr in output.user_attr : attr.is_editable == false && attr.is_list == false && attr.is_visible == true
    ])
    error_message = jsonencode(output.user_attr)
  }
}

run "test_user_attr_change" {
  assert {
    condition     = output.user_attr_change.is_editable == true && output.user_attr_change.is_list == true && output.user_attr_change.is_visible == false
    error_message = jsonencode(output.user_attr)
  }
}
