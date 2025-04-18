---
page_title: "lldap_group Data Source - terraform-provider-lldap"
description: |-
  Reads a LLDAP group, including memberships
---

# lldap_group (Data Source)

Reads a LLDAP group, including memberships

## Example Usage

```terraform
# Get the default admin group "lldap_admin"
data "lldap_group" "lldap_admin" {
  id = 1
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (Number) The unique group ID

### Read-Only

- `attributes` (Set of Object) Attributes for this group (see [below for nested schema](#nestedatt--attributes))
- `creation_date` (String) Metadata of group object creation
- `display_name` (String) Display name of this group
- `users` (Set of Object) Members of this group (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--attributes"></a>
### Nested Schema for `attributes`

Read-Only:

- `name` (String)
- `value` (Set of String)


<a id="nestedatt--users"></a>
### Nested Schema for `users`

Read-Only:

- `display_name` (String)
- `id` (String)
