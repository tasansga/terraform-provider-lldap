---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "lldap_group_memberships Resource - terraform-provider-lldap"
subcategory: ""
description: |-
  Exclusively manages all LLDAP memberhips for this specific group
---

# lldap_group_memberships (Resource)

Exclusively manages all LLDAP memberhips for this specific group

## Example Usage

```terraform
resource "lldap_group" "group" {
  display_name = "Test group"
}

resource "lldap_user" "user" {
  count        = 10
  username     = "user-${count.index}"
  email        = "user-${count.index}@this.test"
}

resource "lldap_group_memberships" "group" {
  group_id = lldap_group.group.id
  user_ids = toset([ for user in lldap_user.user : user.id ])
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group_id` (String) The unique group id
- `user_ids` (Set of String) User ids that must be members of this group

### Read-Only

- `id` (String) ID representing this specific group memberships

## Import

Import is supported using the following syntax:

```shell
# A group's memberships can be imported by specifying the group ID.
terraform import lldap_group_memberships.example 3
```
