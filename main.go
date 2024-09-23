package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	lldap "github.com/tasansga/tf-provider-lldap/lldap"
)

func main() {
	if os.Getenv("DEBUG_LOCAL") == "yes" {
		client := lldap.LldapClient{
			Config: &lldap.Config{
				Url:      "http://localhost:17170",
				UserName: "admin",
				Password: "this_is_a_very_safe_password",
			},
			Token: "",
		}
		diags := client.Authenticate()
		// List groups
		groups, getGroupsErr := client.GetGroups()
		if getGroupsErr != nil {
			fmt.Println(getGroupsErr)
		}
		for _, group := range groups {
			fmt.Println(group.DisplayName)
		}
		// List users
		users, getUsersErr := client.GetUsers()
		if getUsersErr != nil {
			fmt.Println(getUsersErr)
		}
		for _, user := range users {
			fmt.Println(user.DisplayName)
		}
		for _, diag := range diags {
			fmt.Println(diag.Summary)
		}
	} else {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: lldap.Provider,
		})
	}
}
