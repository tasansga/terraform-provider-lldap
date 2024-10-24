run "test_member" {
    assert {
        condition = jsonencode(output.test) == jsonencode({
            "id":"4:admin",
            "group_id":4,
            "user_id":"admin"
            "group_display_name":"Test member group",
        })
        error_message = jsonencode(output.test)
    }
}
