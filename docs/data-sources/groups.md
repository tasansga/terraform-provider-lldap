---
page_title: "lldap_groups Data Source - terraform-provider-lldap"
description: |-
  Reads all LLDAP groups, without memberships
---

# lldap_groups (Data Source)

Reads all LLDAP groups, without memberships

## Example Usage

```terraform
# This data source does not require any input variables
data "lldap_groups" "groups" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `groups` (Set of Object) Set of all groups (see [below for nested schema](#nestedatt--groups))
- `id` (String) Generated ID representing the groups

<a id="nestedatt--groups"></a>
### Nested Schema for `groups`

Read-Only:

- `creation_date` (String)
- `display_name` (String)
- `id` (String)