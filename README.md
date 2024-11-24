# terraform-provider-lldap

Welcome to this unofficial terraform provider for [LLDAP](https://github.com/lldap/lldap/)!

You can find it in the [terraform registry](https://registry.terraform.io/providers/tasansga/lldap/) and in the [OpenTofu registry](https://search.opentofu.org/provider/tasansga/lldap/latest).

[Homepage on GitHub](https://github.com/tasansga/terraform-provider-lldap)


## Features

User, group and membership lifecycle management works and most attributes can be defined in their respective resource. Passwords can be set and changed (but not read). Custom attributes are supported as well.


## Usage

Check the [docs](./docs/index.md)!


## Develop

Just run `make` in the repository root, this will lint, build, test, run `go mod tidy`
and generate docs.

Works for me with:
- Go 1.23
- GNU make 4.4
- Bash 5
- LLDAP 0.6.0
- Docker 27.3

The client uses mostly the [LLDAP GraphQL API](https://github.com/lldap/lldap/blob/main/docs/scripting.md) (see also the [schema](https://github.com/lldap/lldap/blob/main/schema.graphql)).

Password changes use the [password modify extended operation](https://datatracker.ietf.org/doc/html/rfc3062).


## License

See [LICENSE](./LICENSE), or tl;dr:

```
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
```
