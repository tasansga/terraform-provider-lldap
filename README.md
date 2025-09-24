# terraform-provider-lldap

Welcome to this unofficial terraform provider for [LLDAP](https://github.com/lldap/lldap/)!

You can find it in the [terraform registry](https://registry.terraform.io/providers/tasansga/lldap/) and in the [OpenTofu registry](https://search.opentofu.org/provider/tasansga/lldap/latest).

[Homepage on GitHub](https://github.com/tasansga/terraform-provider-lldap)


## Features

User, group and membership lifecycle management works and most attributes can be defined in their respective resource. Passwords can be set and changed (but not read). Custom attributes are supported as well.


## Usage

Check the [docs](./docs/index.md)!

## lldap-cli

`lldap-cli` is a command line tool for interacting with an LLDAP server.

### Setup

Before using `lldap-cli`, you need to set the following environment variables:

- `LLDAP_USER` (optional, defaults to "admin") - The username of the administrative user.
- `LLDAP_PASSWORD` (required) - Password for the administrative user.
- `LLDAP_BASE_DN` (required) - LDAP Base Distinguished Name, for example: `dc=example,dc=com`.
- `LLDAP_HTTP_URL` (required) - HTTP(s) URL of the LLDAP server, e.g. `https://localhost:3000`.
- `LLDAP_LDAP_URL` (required) - LDAP(s) URL of the LLDAP server, e.g. `ldaps://localhost:636`.
- `INSECURE_CERT` (optional, default `false`) - Skip TLS certificate verification if set to `true`.

### Basic Usage

```bash
$ lldap-cli --help
Usage:
  lldap-cli [command]

Available Commands:
  attribute   Attribute operations
  group       Group operations
  member      Membership operations
  user        User operations
```


## Develop

Just run `make` in the repository root, this will lint, build, test, run `go mod tidy`
and generate docs.

Works for me with:
- Go 1.25
- GNU make 4.4
- Bash 5
- LLDAP 0.6.2
- Docker 27.5

The client uses mostly the [LLDAP GraphQL API](https://github.com/lldap/lldap/blob/main/docs/scripting.md) (see also the [schema](https://github.com/lldap/lldap/blob/main/schema.graphql)).

Password changes use the [password modify extended operation](https://datatracker.ietf.org/doc/html/rfc3062).


## License

See [LICENSE](./LICENSE), or tl;dr:

```
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
```
