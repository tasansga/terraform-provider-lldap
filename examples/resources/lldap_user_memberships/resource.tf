resource "lldap_user" "user" {
  username     = "user"
  email        = "user@this.test"
}

resource "lldap_group" "group" {
  count        = 10
  display_name = "Test group ${count.index}"
}

resource "lldap_user_memberships" "user" {
  user_id = lldap_user.user.id
  group_ids = toset([ for group in lldap_group.group : group.id ])
}
