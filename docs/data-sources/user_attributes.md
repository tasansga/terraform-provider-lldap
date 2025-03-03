---
page_title: "lldap_user_attributes Data Source - terraform-provider-lldap"
description: |-
  Schema definitions for user attributes
---

# lldap_user_attributes (Data Source)

Schema definitions for user attributes

## Example Usage

```terraform
# Read all user attribute schema definitions
data "lldap_user_attributes" "this" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `attributes` (Set of Object) Set of all user attributes (see [below for nested schema](#nestedatt--attributes))
- `id` (String) Generated ID representing the attributes

<a id="nestedatt--attributes"></a>
### Nested Schema for `attributes`

Read-Only:

- `attribute_type` (String)
- `is_editable` (Boolean)
- `is_hardcoded` (Boolean)
- `is_list` (Boolean)
- `is_readonly` (Boolean)
- `is_visible` (Boolean)
- `name` (String)
