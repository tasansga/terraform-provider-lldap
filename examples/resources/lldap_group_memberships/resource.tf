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
