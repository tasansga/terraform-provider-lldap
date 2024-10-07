run "test_group" {
    assert {
        condition = jsonencode(output.test) == jsonencode({
            "creation_date": null,
            "display_name": "Test group",
            "id": "4",
            "users": []
        })
        error_message = jsonencode(output.test)
    }
}