# Settings for connections to a local LLDAP docker container
# `docker run -p 17170 -p 3890 --env "LLDAP_LDAP_USER_PASS=lldap-admin-password"`
provider "lldap" {
  http_url = "http://localhost:17170"
  ldap_url = "ldap://localhost:3890"
  username = "admin"
  password = "lldap-admin-password"
  base_dn  = "dc=example,dc=com"
}
