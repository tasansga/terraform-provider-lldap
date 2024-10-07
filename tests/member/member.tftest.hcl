/*
run "test_member" {
    assert {
        condition = jsonencode(output.test) == jsonencode({
            "creation_date": null,
        })
        error_message = jsonencode(output.test)
    }
}
*/