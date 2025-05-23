/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spf13/cobra"
	lldap "github.com/tasansga/terraform-provider-lldap/lldap"
)

var logger = slog.Default()
var lc lldap.LldapClient

var isInit = false

func getClient(ctx context.Context) (*lldap.LldapClient, error) {
	username := os.Getenv("LLDAP_USER")
	if username == "" {
		logger.Debug("LLDAP_USER not set, defaulting to 'admin'")
		username = "admin"
	}
	password := os.Getenv("LLDAP_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("LLDAP_PASSWORD not set")
	}
	baseDn := os.Getenv("LLDAP_BASE_DN")
	if baseDn == "" {
		return nil, fmt.Errorf("LLDAP_BASE_DN not set")
	}
	rawHttpUrl := os.Getenv("LLDAP_HTTP_URL")
	parsedHttpUrl, parseHttpUrlErr := url.Parse(rawHttpUrl)
	if parseHttpUrlErr != nil {
		return nil, parseHttpUrlErr
	}
	if parsedHttpUrl.Scheme != "http" && parsedHttpUrl.Scheme != "https" {
		return nil, fmt.Errorf("invalid value for LLDAP_HTTP_URL: '%s'", rawHttpUrl)
	}
	rawLdapUrl := os.Getenv("LLDAP_LDAP_URL")
	parsedLdapUrl, parseLdapUrlErr := url.Parse(rawLdapUrl)
	if parseLdapUrlErr != nil {
		return nil, parseLdapUrlErr
	}
	if parsedLdapUrl.Scheme != "ldap" && parsedLdapUrl.Scheme != "ldaps" {
		return nil, fmt.Errorf("invalid value for LLDAP_LDAP_URL: '%s'", rawLdapUrl)
	}
	insecureCertStr := os.Getenv("INSECURE_CERT")
	insecureCert := false
	if insecureCertStr != "" {
		var parseInsecCertErr error
		insecureCert, parseInsecCertErr = strconv.ParseBool(insecureCertStr)
		if parseInsecCertErr != nil {
			return nil, parseInsecCertErr
		}
	}
	client := lldap.LldapClient{
		Config: lldap.Config{
			Context:               ctx,
			HttpUrl:               parsedHttpUrl,
			LdapUrl:               parsedLdapUrl,
			UserName:              username,
			Password:              password,
			BaseDn:                baseDn,
			InsecureSkipCertCheck: insecureCert,
		},
	}
	return &client, nil
}

var rootCmd = &cobra.Command{
	Use:   "lldap-cli",
	Short: "Basic client CLI to interact with a LLDAP server",
	Long: `lldap-cli is a command line tool for interacting with an LLDAP server.

The following environment variables are supported by all subcommands:
- LLDAP_USER      (optional, default: 'admin', username for the administrative user)
- LLDAP_PASSWORD  (required, password for the administrative user)
- LLDAP_BASE_DN   (required, LDAP base DN in the format 'dc=example,dc=com')
- LLDAP_HTTP_URL  (required, HTTP URL in the format 'http[s]://(hostname)[:port]')
- LLDAP_LDAP_URL  (required, LDAP URL in the format 'ldap[s]://(hostname)[:port]')
- INSECURE_CERT   (optional, default: 'false', skip cert check for HTTPS connections)`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.CalledAs() == "help" || cmd.Flags().Lookup("help").Changed {
			return nil
		}
		lclient, getClientErr := getClient(cmd.Context())
		if getClientErr != nil {
			return getClientErr
		}
		lc = *lclient
		return nil
	},
}

var userCmds = map[string]*cobra.Command{
	"create": {
		Use:           "create <username>",
		Short:         "Create a new user",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			user := lldap.LldapUser{
				Id: args[0],
			}
			if email, err := cmd.Flags().GetString("email"); err == nil && email != "" {
				user.Email = email
			}
			if displayName, err := cmd.Flags().GetString("displayname"); err == nil && displayName != "" {
				user.DisplayName = displayName
			}
			if firstName, err := cmd.Flags().GetString("firstname"); err == nil && firstName != "" {
				user.FirstName = firstName
			}
			if lastName, err := cmd.Flags().GetString("lastname"); err == nil && lastName != "" {
				user.LastName = lastName
			}
			if avatar, err := cmd.Flags().GetString("avatar"); err == nil && avatar != "" {
				user.Avatar = avatar
			}
			createErr := lc.CreateUser(&user)
			if createErr != nil {
				logger.Error("could not create user", slog.Any("error", createErr))
				return fmt.Errorf("could not create user")
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(user)
		},
	},
	"get": {
		Use:           "get [uid]",
		Short:         "Get a list of all users or details of a given user",
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				user, getErr := lc.GetUser(args[0])
				if getErr != nil {
					logger.Error("could not get user", slog.Any("error", getErr))
					return fmt.Errorf("could not get user")
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(user)
			}
			users, getErr := lc.GetUsers()
			if getErr != nil {
				logger.Error("could not get users", slog.Any("error", getErr))
				return fmt.Errorf("could not get users")
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(users)
		},
	},
	"update": {
		Use:           "update <uid>",
		Short:         "Update an user",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			user, getErr := lc.GetUser(args[0])
			if getErr != nil {
				logger.Error("could not get user - invalid user id?", slog.Any("error", getErr))
				return fmt.Errorf("could not get user - invalid user id?")
			}
			if email, err := cmd.Flags().GetString("email"); err == nil && email != "" {
				user.Email = email
			}
			if displayName, err := cmd.Flags().GetString("displayname"); err == nil && displayName != "" {
				user.DisplayName = displayName
			}
			if firstName, err := cmd.Flags().GetString("firstname"); err == nil && firstName != "" {
				user.FirstName = firstName
			}
			if lastName, err := cmd.Flags().GetString("lastname"); err == nil && lastName != "" {
				user.LastName = lastName
			}
			if avatar, err := cmd.Flags().GetString("avatar"); err == nil && avatar != "" {
				user.Avatar = avatar
			}
			updateErr := lc.UpdateUser(user)
			if updateErr != nil {
				logger.Error("could not update user", slog.Any("error", updateErr), slog.Any("user", user))
				return fmt.Errorf("could not update user")
			}
			logger.Info("user updated", slog.Any("user", user))
			return nil
		},
	},
	"delete": {
		Use:           "delete <uid>",
		Short:         "Delete an user",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			delErr := lc.DeleteUser(args[0])
			if delErr != nil {
				logger.Error("could not delete user", slog.Any("error", delErr), slog.String("uid", args[0]))
				return fmt.Errorf("could not delete user")
			}
			logger.Info("user deleted", slog.String("uid", args[0]))
			return nil
		},
	},
	"password": {
		Use:           "password <uid> <password>",
		Short:         "Update the password of an user",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			pwdErr := lc.SetUserPassword(args[0], args[1])
			if pwdErr != nil {
				logger.Error("could not set user password", slog.Any("error", pwdErr))
				return fmt.Errorf("could not set user password")
			}
			logger.Info("user password changed", slog.String("uid", args[0]))
			return nil
		},
	},
}

var groupCmds = map[string]*cobra.Command{
	"create": {
		Use:           "create <name>",
		Short:         "Create a new group",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			displayName := args[0]
			if flagDisplayName, err := cmd.Flags().GetString("displayname"); err == nil && flagDisplayName != "" {
				displayName = flagDisplayName
			}
			group := lldap.LldapGroup{
				DisplayName: displayName,
			}
			createErr := lc.CreateGroup(&group)
			if createErr != nil {
				logger.Error("could not create group", slog.Any("error", createErr))
				return fmt.Errorf("could not create group")
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(group)
		},
	},
	"get": {
		Use:           "get [gid]",
		Short:         "Get a list of all groups or details of a given groups",
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				gid, invalidGid := strconv.Atoi(args[0])
				if invalidGid != nil {
					logger.Error("invalid gid", slog.Any("err", invalidGid))
					return fmt.Errorf("invalid gid: %w", invalidGid)
				}
				group, getErr := lc.GetGroup(gid)
				if getErr != nil {
					logger.Error("could not get group", slog.Any("err", getErr))
					return fmt.Errorf("could not get group")
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(group)
			}
			groups, getErr := lc.GetGroups()
			if getErr != nil {
				logger.Error("could not get groups", slog.Any("err", getErr))
				return fmt.Errorf("could not get groups")
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(groups)
		},
	},
	"update": {
		Use:           "update <gid>",
		Short:         "Update a group",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			gid, invalidGid := strconv.Atoi(args[0])
			if invalidGid != nil {
				logger.Error("invalid gid", slog.Any("err", invalidGid))
				return invalidGid
			}
			displayName, err := cmd.Flags().GetString("displayname")
			if err != nil {
				return err
			}
			updateErr := lc.UpdateGroupDisplayName(gid, displayName)
			if updateErr != nil {
				logger.Error("could not update group", slog.Any("error", updateErr))
				return fmt.Errorf("could not update user")
			}
			logger.Info("group displayname updated", slog.Any("displayname", displayName))
			return nil
		},
	},
	"delete": {
		Use:           "delete <gid>",
		Short:         "Delete a group",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			gid, invalidGid := strconv.Atoi(args[0])
			if invalidGid != nil {
				logger.Error("invalid gid", slog.Any("err", invalidGid))
				return invalidGid
			}
			delErr := lc.DeleteGroup(gid)
			if delErr != nil {
				logger.Error("could not delete group", slog.Any("err", delErr))
				return fmt.Errorf("could not delete group")
			}
			logger.Info("group deleted", slog.String("gid", args[0]))
			return nil
		},
	},
}

var memberCmds = map[string]*cobra.Command{
	"add": {
		Use:           "add <gid> <uid>",
		Short:         "Add a member to a group",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			gid, invalidGid := strconv.Atoi(args[0])
			if invalidGid != nil {
				logger.Error("invalid gid", slog.Any("err", invalidGid))
				return invalidGid
			}
			uid := args[1]
			addErr := lc.AddUserToGroup(gid, uid)
			if addErr != nil {
				logger.Error("could not add user to group", slog.Any("err", addErr))
				return fmt.Errorf("could not add user to group")
			}
			logger.Info("added member to group", slog.Int("gid", gid), slog.String("uid", uid))
			return nil
		},
	},
	"remove": {
		Use:           "remove <gid> <uid>",
		Short:         "Remove a member from a group",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			gid, invalidGid := strconv.Atoi(args[0])
			if invalidGid != nil {
				logger.Error("invalid gid", slog.Any("err", invalidGid))
				return invalidGid
			}
			uid := args[1]
			addErr := lc.RemoveUserFromGroup(gid, uid)
			if addErr != nil {
				logger.Error("could not remove user from group", slog.Any("err", addErr))
				return fmt.Errorf("could not remove user from group")
			}
			logger.Info("removed member from group", slog.Int("gid", gid), slog.String("uid", uid))
			return nil
		},
	},
}

var attributeCmds = map[string]*cobra.Command{
	"create": {
		Use:           "create <name> <type> (--user | --group)",
		Short:         "Create a new custom attribute schema for users or groups",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if !slices.Contains(lldap.VALID_ATTRIBUTE_TYPES, strings.ToUpper(args[1])) {
				return fmt.Errorf("invalid attribute type %s, expected one of %s", args[1], lldap.VALID_ATTRIBUTE_TYPES)
			}
			attributeType := lldap.LldapCustomAttributeType(strings.ToUpper(args[1]))
			flagUser, _ := cmd.Flags().GetBool("user")
			flagGroup, _ := cmd.Flags().GetBool("group")
			if !flagUser && !flagGroup {
				return fmt.Errorf("either --user or --group must be set")
			}
			flagIsList, _ := cmd.Flags().GetBool("list")
			flagIsVisible, _ := cmd.Flags().GetBool("visible")
			flagIsEditable, _ := cmd.Flags().GetBool("editable")
			displayName, _ := cmd.Flags().GetString("displayname")
			var userAttr *lldap.LldapUserAttributeSchema
			var groupAttr *lldap.LldapGroupAttributeSchema
			var createErr diag.Diagnostics
			var getErr diag.Diagnostics
			if flagUser {
				createErr = lc.CreateUserAttribute(
					name,
					attributeType,
					flagIsList,
					flagIsVisible,
					flagIsEditable,
				)
				userAttr, getErr = lc.GetUserAttributeSchema(name)
			}
			if flagGroup {
				createErr = lc.CreateGroupAttribute(
					name,
					attributeType,
					flagIsList,
					flagIsVisible,
				)
				groupAttr, getErr = lc.GetGroupAttributeSchema(name)
			}
			if createErr != nil {
				logger.Error("could not create attribute", slog.Any("err", createErr))
				return fmt.Errorf("could not create attribute")
			}
			if getErr != nil {
				logger.Error("could not get created attribute", slog.Any("err", createErr))
				return fmt.Errorf("could not get created attribute")
			}
			if displayName != "" {
				logger.Info("attribute created", slog.String("name", name), slog.String("type", string(attributeType)), slog.String("displayname", displayName))
			} else {
				logger.Info("attribute created", slog.String("name", name), slog.String("type", string(attributeType)))
			}
			if userAttr != nil {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(userAttr)
			}
			if groupAttr != nil {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(groupAttr)
			}
			return nil
		},
	},
	"add": {
		Use:           "add <attribute> <userOrGroupId> (--user | --group)",
		Short:         "Add a custom attribute to an user or a group",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			attribute := args[0]
			id := args[1]
			flagUser, _ := cmd.Flags().GetBool("user")
			flagGroup, _ := cmd.Flags().GetBool("group")
			if !flagUser && !flagGroup {
				return fmt.Errorf("either --user or --group must be set")
			}
			flagValues, _ := cmd.Flags().GetStringSlice("values")
			var addErr diag.Diagnostics
			if flagUser {
				addErr = lc.AddAttributeToUser(id, attribute, flagValues)
			}
			if flagGroup {
				gid, invalidGid := strconv.Atoi(id)
				if invalidGid != nil {
					logger.Error("invalid gid", slog.Any("err", invalidGid))
					return invalidGid
				}
				addErr = lc.AddAttributeToGroup(gid, attribute, flagValues)
			}
			if addErr != nil {
				logger.Error("could not add attribute",
					slog.Any("err", addErr),
					slog.Any("values", flagValues),
					slog.String("id", id),
					slog.String("attribute", attribute),
				)
				return fmt.Errorf("could not add attribute")
			}
			logger.Info("attribute added",
				slog.Any("values", flagValues),
				slog.String("id", id),
				slog.String("attribute", attribute),
			)
			return nil
		},
	},
	"delete": {
		Use:           "delete <attribute> (--user | --group)",
		Short:         "Delete a custom attribute schema",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flagUser, _ := cmd.Flags().GetBool("user")
			flagGroup, _ := cmd.Flags().GetBool("group")
			var addErr diag.Diagnostics
			if flagUser {
				addErr = lc.DeleteUserAttribute(args[0])
			}
			if flagGroup {
				addErr = lc.DeleteGroupAttribute(args[0])
			}
			if addErr != nil {
				logger.Error("could not delete attribute",
					slog.String("attribute", args[0]),
				)
				return fmt.Errorf("could not delete attribute")
			}
			slog.Info("attribute deleted", slog.String("attribute", args[0]))
			return nil
		},
	},
	"remove": {
		Use:           "remove <attribute> <userOrGroupId> (--user | --group)",
		Short:         "Remove a custom attribute from an user or a group",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flagUser, _ := cmd.Flags().GetBool("user")
			flagGroup, _ := cmd.Flags().GetBool("group")
			var removeErr diag.Diagnostics
			if flagUser {
				removeErr = lc.RemoveAttributeFromUser(args[1], args[0])
			}
			if flagGroup {
				gid, invalidGid := strconv.Atoi(args[1])
				if invalidGid != nil {
					logger.Error("invalid gid", slog.Any("err", invalidGid))
					return invalidGid
				}
				removeErr = lc.RemoveAttributeFromGroup(gid, args[0])
			}
			if removeErr != nil {
				logger.Error("could not remove attribute",
					slog.String("attribute", args[0]),
					slog.String("id", args[1]),
				)
				return fmt.Errorf("could not remove attribute")
			}
			slog.Info("attribute removed", slog.String("attribute", args[0]), slog.String("id", args[1]))
			return nil
		},
	},
	"schema": {
		Use:           "schema [name]",
		Short:         "Get all custom attribute schemas or for a specific user/group attribute",
		Args:          cobra.RangeArgs(0, 1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flagUser, _ := cmd.Flags().GetBool("user")
			flagGroup, _ := cmd.Flags().GetBool("group")

			var result any
			getErr := diag.FromErr(fmt.Errorf("invalid value for user|group argument"))
			if flagUser {
				if len(args) == 1 {
					result, getErr = lc.GetUserAttributeSchema(args[0])
				} else {
					result, getErr = lc.GetUserAttributesSchema()
				}
			}
			if flagGroup {
				if len(args) == 1 {
					result, getErr = lc.GetGroupAttributeSchema(args[0])
				} else {
					result, getErr = lc.GetGroupAttributesSchema()
				}
			}
			if getErr != nil {
				logger.Error("could not get attribute", slog.Any("err", getErr))
				return fmt.Errorf("could not get attribute")
			}
			if result != nil {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
			}
			return nil
		},
	},
}

var mainCmds = map[string]*cobra.Command{
	"user": {
		Use:   "user",
		Short: "User operations",
	},
	"group": {
		Use:   "group",
		Short: "Group operations",
	},
	"member": {
		Use:   "member",
		Short: "Membership operations",
	},
	"attribute": {
		Use:   "attribute",
		Short: "Attribute operations",
	},
}

func initCmds() *cobra.Command {
	if isInit {
		return rootCmd
	}
	for _, cmdName := range []string{"create", "update"} {
		userCmds[cmdName].Flags().String("displayname", "", "Display name")
		userCmds[cmdName].Flags().String("email", "", "Email")
		userCmds[cmdName].Flags().String("firstname", "", "First name")
		userCmds[cmdName].Flags().String("lastname", "", "Last name")
		userCmds[cmdName].Flags().String("avatar", "", "Base 64 encoded JPEG image")
	}
	groupCmds["create"].Flags().String("displayname", "", "Display name")
	groupCmds["update"].Flags().String("displayname", "", "Display name")
	attributeCmds["create"].Flags().Bool("list", false, "Does this attribute represent a list?")
	attributeCmds["create"].Flags().Bool("visible", true, "Is this attribute visible in LDAP?")
	attributeCmds["create"].Flags().Bool("editable", false, "Is this attribute user editable?")
	attributeCmds["create"].Flags().String("displayname", "", "Display name")
	attributeCmds["add"].Flags().StringSlice("values", nil, "List of values for this attribute")
	for _, cmd := range userCmds {
		mainCmds["user"].AddCommand(cmd)
	}
	for _, cmd := range groupCmds {
		mainCmds["group"].AddCommand(cmd)
	}
	for _, cmd := range memberCmds {
		mainCmds["member"].AddCommand(cmd)
	}
	for _, cmd := range attributeCmds {
		cmd.Flags().Bool("user", false, "Handle user-specific attribute")
		cmd.Flags().Bool("group", false, "Handle group-specific attribute")
		mainCmds["attribute"].AddCommand(cmd)
	}
	for _, cmd := range mainCmds {
		rootCmd.AddCommand(cmd)
	}
	isInit = true
	return rootCmd
}

func main() {
	cmd := initCmds()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
