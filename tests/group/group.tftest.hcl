run "test_group" {
  assert {
    condition     = output.test.id != null && output.test.id != 0
    error_message = "id should not be null or zero"
  }
  assert {
    condition     = output.test.display_name == "Test group"
    error_message = "display_name should be 'Test group'"
  }
  assert {
    condition     = output.test.creation_date != null && output.test.creation_date != ""
    error_message = "creation_date should not be null or empty string"
  }
  assert {
    condition     = output.test.uuid != null && output.test.uuid != ""
    error_message = "uuid should not be null or empty string"
  }
  assert {
    condition     = output.test.users != null
    error_message = "users should not be null"
  }
}
