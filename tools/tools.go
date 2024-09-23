//go:build tools

package tools

import (
	// Documentation generation
	// https://github.com/hashicorp/terraform-plugin-docs
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
