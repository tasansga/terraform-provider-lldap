---
page_title: "lldap_user Resource - terraform-provider-lldap"
description: |-
  Manages a LLDAP user
---

# lldap_user (Resource)

Manages a LLDAP user

## Example Usage

```terraform
# Manage an user with attributes
resource "lldap_user" "user" {
  username     = "myuser"
  email        = "myuser@in.the.test"
  password     = "super-secret password!"
  display_name = "My User"
  first_name   = "My"
  last_name    = "User"
  avatar       = filebase64("${path.module}/myuser.jpeg")
}
```

## Import

Import is supported using the following syntax:

```sh
# An user can be imported by specifying the user ID (username).
terraform import lldap_user.example admin
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email` (String) The unique user email
- `username` (String) The unique username

### Optional

- `avatar` (String) Base 64 encoded JPEG image
- `display_name` (String) Display name of this user
- `first_name` (String) First name of this user
- `last_name` (String) Last name of this user
- `password` (String, Sensitive) Password for the user

### Read-Only

- `creation_date` (String) Metadata of user object creation
- `id` (String) ID representing this specific user
- `uuid` (String) UUID of user
