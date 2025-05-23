//go:build integration
// +build integration

/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/exp/rand"

	"github.com/stretchr/testify/assert"
	lldap "github.com/tasansga/terraform-provider-lldap/lldap"
)

// getTestClient creates a real LLDAP client for integration tests
func getTestClient() lldap.LldapClient {
	hostIp := os.Getenv("LLDAP_HOST")
	password := os.Getenv("LLDAP_PASSWORD")
	httpPort := os.Getenv("LLDAP_PORT_HTTP")
	ldapPort := os.Getenv("LLDAP_PORT_LDAP")
	parsedHttpUrl, _ := url.Parse(fmt.Sprintf("http://%s:%s", hostIp, httpPort))
	parsedLdapUrl, _ := url.Parse(fmt.Sprintf("ldap://%s:%s", hostIp, ldapPort))
	client := lldap.LldapClient{
		Config: lldap.Config{
			HttpUrl:  parsedHttpUrl,
			LdapUrl:  parsedLdapUrl,
			UserName: "admin",
			Password: password,
			BaseDn:   "dc=example,dc=com",
		},
	}
	return client
}

// randomTestSuffix generates a random suffix for test names
func randomTestSuffix(s string) string {
	r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	anums := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = anums[r.Intn(len(anums))]
	}
	return fmt.Sprintf("%s-%s", s, string(b))
}

// integrationTestWrap executes a command with the given arguments and returns the output
// This is separate from the unit test testWrap to avoid conflicts
func integrationTestWrap(args []string) (*bytes.Buffer, *bytes.Buffer, error) {
	// Reset all command flags to avoid "flag redefined" errors
	if isInit {
		// Reset flags on all commands that have them
		for _, cmdName := range []string{"create", "update"} {
			if cmd, exists := userCmds[cmdName]; exists {
				cmd.ResetFlags()
			}
		}
		if cmd, exists := groupCmds["create"]; exists {
			cmd.ResetFlags()
		}
		if cmd, exists := groupCmds["update"]; exists {
			cmd.ResetFlags()
		}
		for _, cmd := range attributeCmds {
			cmd.ResetFlags()
		}

		// Reset root command
		rootCmd.ResetFlags()
		rootCmd.ResetCommands()

		// Reset isInit so commands can be re-initialized
		isInit = false
	}

	cmd := initCmds()
	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)
	cmd.SetOut(stdOut)
	cmd.SetErr(stdErr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdOut, stdErr, err
}

// Integration tests start here

func TestUserPasswordCommand(t *testing.T) {
	username := randomTestSuffix("testpassword")
	email := strings.Join([]string{username, "test.local"}, "@")
	client := getTestClient()

	// Create a test user first
	testUser := lldap.LldapUser{
		Id:    username,
		Email: email,
	}
	cerr := client.CreateUser(&testUser)
	assert.Nil(t, cerr)

	// Test password setting
	stdOut, stdErr, err := integrationTestWrap([]string{
		"user",
		"password",
		username,
		"newpassword123",
	})

	assert.Empty(t, stdOut) // Password command should not output anything on success
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	// Clean up
	client.DeleteUser(username)
}

func TestGroupGetAll(t *testing.T) {
	// Test getting all groups
	stdOut, stdErr, err := integrationTestWrap([]string{
		"group",
		"get",
	})

	assert.NotNil(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	// Parse JSON output
	var groups []lldap.LldapGroup
	err = json.Unmarshal(stdOut.Bytes(), &groups)
	assert.Nil(t, err)
	assert.NotEmpty(t, groups)
	for _, group := range groups {
		assert.NotEqual(t, 0, group.Id)
		assert.NotEmpty(t, group.DisplayName)
	}
}

func TestMemberRemove(t *testing.T) {
	username := randomTestSuffix("testmemberremoveuser")
	client := getTestClient()

	// Create test group
	testGroup := lldap.LldapGroup{
		DisplayName: "Test Member Remove Group",
	}
	derr := client.CreateGroup(&testGroup)
	assert.Nil(t, derr)

	// Create test user
	derr = client.CreateUser(&lldap.LldapUser{Id: username, Email: username + "@test.local"})
	assert.Nil(t, derr)

	// Add user to group first
	derr = client.AddUserToGroup(testGroup.Id, username)
	assert.Nil(t, derr)

	// Test removing member via CLI
	stdOut, stdErr, err := integrationTestWrap([]string{
		"member",
		"remove",
		fmt.Sprint(testGroup.Id),
		username,
	})

	assert.Empty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	// Verify member was removed
	updatedGroup, derr := client.GetGroup(testGroup.Id)
	assert.Nil(t, derr)

	found := false
	for _, member := range updatedGroup.Users {
		if member.Id == username {
			found = true
			break
		}
	}
	assert.False(t, found, "Member should have been removed")

	// Clean up
	client.DeleteUser(username)
	client.DeleteGroup(testGroup.Id)
}

func TestAttributeSchemaUser(t *testing.T) {

	// Test getting all user attribute schemas
	stdOut, stdErr, err := integrationTestWrap([]string{
		"attribute",
		"schema",
		"--user",
	})

	assert.NotEmpty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	// Parse JSON output
	var attrs []lldap.LldapUserAttributeSchema
	err = json.Unmarshal(stdOut.Bytes(), &attrs)
	assert.Nil(t, err)
	assert.NotEmpty(t, attrs)

	// Check for some expected standard attributes (using LLDAP internal names)
	expectedAttributes := []string{"mail", "display_name", "first_name", "last_name"}
	for _, expected := range expectedAttributes {
		found := false
		for _, attr := range attrs {
			if attr.Name == expected {
				found = true
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("Expected attribute %s not found", expected))
	}
}

func TestAttributeSchemaGroup(t *testing.T) {

	// Test getting all group attribute schemas
	stdOut, stdErr, err := integrationTestWrap([]string{
		"attribute",
		"schema",
		"--group",
	})

	assert.NotEmpty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	// Parse JSON output
	var attrs []lldap.LldapGroupAttributeSchema
	err = json.Unmarshal(stdOut.Bytes(), &attrs)
	assert.Nil(t, err)
	assert.NotEmpty(t, attrs)

	// Verify structure
	for _, attr := range attrs {
		assert.NotEmpty(t, attr.Name)
		assert.NotEmpty(t, attr.AttributeType)
	}
}

func TestAttributeSchemaSpecific(t *testing.T) {

	attrName := randomTestSuffix("testschemaattr")
	client := getTestClient()

	// Create a test attribute first
	derr := client.CreateUserAttribute(attrName, lldap.LldapCustomAttributeType("STRING"), false, true, true)
	assert.Nil(t, derr)

	// Test getting specific attribute schema
	stdOut, stdErr, err := integrationTestWrap([]string{
		"attribute",
		"schema",
		"--user",
		attrName,
	})

	assert.NotEmpty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	// Parse JSON output
	var attr lldap.LldapUserAttributeSchema
	err = json.Unmarshal(stdOut.Bytes(), &attr)
	assert.Nil(t, err)
	assert.Equal(t, attrName, attr.Name)
	assert.Equal(t, lldap.LldapCustomAttributeType("STRING"), attr.AttributeType)
	assert.True(t, attr.IsVisible)
	assert.True(t, attr.IsEditable)
	assert.False(t, attr.IsList)

	// Clean up
	client.DeleteUserAttribute(attrName)
}

func TestUserGetInvalidUser(t *testing.T) {

	// Test getting a non-existent user
	nonExistentUser := "user-does-not-exist-" + randomTestSuffix("test")
	stdOut, stdErr, err := integrationTestWrap([]string{
		"user",
		"get",
		nonExistentUser,
	})

	assert.Empty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "could not get user")
}

func TestGroupGetInvalidGroup(t *testing.T) {

	// Test getting a non-existent group
	stdOut, stdErr, err := integrationTestWrap([]string{
		"group",
		"get",
		"99999", // Very unlikely to exist
	})

	assert.Empty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "could not get group")
}

func TestAttributeCreateInvalidType(t *testing.T) {

	// Test attribute create with invalid type
	attrName := randomTestSuffix("testinvalidtype")
	stdOut, _, err := integrationTestWrap([]string{
		"attribute",
		"create",
		attrName,
		"INVALID_TYPE",
		"--user",
		"--displayname",
		"Test Invalid Type",
	})

	assert.Empty(t, stdOut)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid attribute type")
}

func TestAttributeCreateMissingFlag(t *testing.T) {

	// Test attribute create without --user or --group flag
	attrName := randomTestSuffix("testmissingflag")
	_, _, err := integrationTestWrap([]string{
		"attribute",
		"create",
		attrName,
		"string",
		"--displayname",
		"Test Missing Flag",
	})

	// This should fail with validation error
	assert.NotNil(t, err, "Expected validation error for missing --user or --group flag")
}

// CRUD Integration Tests

func TestUserCreate(t *testing.T) {

	username := randomTestSuffix("testusercreate")
	email := strings.Join([]string{username, "local.test"}, "@")
	stdOut, stdErr, err := integrationTestWrap([]string{
		"user",
		"create",
		username,
		"--email",
		email,
		"--displayname",
		"Test User",
		"--firstname",
		"Test",
		"--lastname",
		"User",
	})

	assert.NotEmpty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	var result lldap.LldapUser
	err = json.Unmarshal(stdOut.Bytes(), &result)
	assert.Nil(t, err)
	assert.Equal(t, username, result.Id)
	assert.Equal(t, email, result.Email)

	// Verify expected attributes are present (using LLDAP internal names)
	expectedAttributes := []string{"mail", "display_name", "first_name", "last_name"}
	for _, expected := range expectedAttributes {
		found := false
		for _, attr := range result.Attributes {
			if attr.Name == expected {
				found = true
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("Expected attribute %s not found", expected))
	}
}

func TestUserGetAll(t *testing.T) {

	stdOut, stdErr, err := integrationTestWrap([]string{
		"user",
		"get",
	})

	assert.NotNil(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	var users []lldap.LldapUser
	err = json.Unmarshal(stdOut.Bytes(), &users)
	assert.Nil(t, err)
	assert.NotEmpty(t, users)
	for _, user := range users {
		assert.NotEmpty(t, user.Id)
	}
}

func TestUserGetOne(t *testing.T) {

	username := randomTestSuffix("testusergetone")
	email := strings.Join([]string{username, "test.local"}, "@")
	client := getTestClient()
	testUser := lldap.LldapUser{
		Id:          username,
		Email:       email,
		DisplayName: "Test User", // Add display name so display_name attribute is present
		FirstName:   "Test",      // Add first name so first_name attribute is present
		LastName:    "User",      // Add last name so last_name attribute is present
	}
	cerr := client.CreateUser(&testUser)
	assert.Nil(t, cerr)

	stdOut, stdErr, err := integrationTestWrap([]string{
		"user",
		"get",
		username,
	})

	assert.NotNil(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)
	var result lldap.LldapUser
	err = json.Unmarshal(stdOut.Bytes(), &result)
	assert.Nil(t, err)
	assert.Equal(t, username, result.Id)

	// Verify expected attributes are present (using LLDAP internal names)
	expectedAttributes := []string{"mail", "display_name", "first_name", "last_name"}
	for _, expected := range expectedAttributes {
		found := false
		for _, attr := range result.Attributes {
			if attr.Name == expected {
				found = true
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("Expected attribute %s not found", expected))
	}
}

func TestUserUpdate(t *testing.T) {

	username := randomTestSuffix("testuserupdate")
	flagEmail := strings.Join([]string{username, "local.test"}, "@")
	flagDisplayName := "Display Name"
	flagFirstName := "First"
	flagLastName := "Last"
	// Use empty avatar to avoid base64 validation issues
	flagAvatar := ""

	testUser := lldap.LldapUser{
		Id:          username,
		Email:       flagEmail,
		DisplayName: flagDisplayName,
		FirstName:   flagFirstName,
		LastName:    flagLastName,
		Avatar:      flagAvatar,
	}
	client := getTestClient()
	cerr := client.CreateUser(&testUser)
	assert.Nil(t, cerr)

	updatedDisplayName := "Updated Display Name"
	flagDisplayName = updatedDisplayName
	stdOut, stdErr, err := integrationTestWrap([]string{
		"user",
		"update",
		testUser.Id,
		"--displayname", flagDisplayName,
	})

	assert.Empty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	stdOut, stdErr, err = integrationTestWrap([]string{
		"user",
		"get",
		testUser.Id,
	})

	assert.NotEmpty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	var updatedUser lldap.LldapUser
	err = json.Unmarshal(stdOut.Bytes(), &updatedUser)
	assert.Nil(t, err)
	assert.Equal(t, testUser.Id, updatedUser.Id)
	assert.Equal(t, flagEmail, updatedUser.Email)
	assert.Equal(t, flagDisplayName, updatedUser.DisplayName)
	assert.Equal(t, flagFirstName, updatedUser.FirstName)
	assert.Equal(t, flagLastName, updatedUser.LastName)
	assert.Equal(t, flagAvatar, updatedUser.Avatar)
}

func TestUserDelete(t *testing.T) {

	username := randomTestSuffix("testuserdelete")
	email := strings.Join([]string{username, "test.local"}, "@")
	client := getTestClient()
	testUser := lldap.LldapUser{
		Id:    username,
		Email: email,
	}
	derr := client.CreateUser(&testUser)
	assert.Nil(t, derr)

	stdOut, stdErr, err := integrationTestWrap([]string{
		"user",
		"delete",
		username,
	})

	assert.NotNil(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	// Verify user was deleted by trying to get it
	_, diags := client.GetUser(username)

	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "Entity not found")
}

func TestGroupCreate(t *testing.T) {

	groupname := randomTestSuffix("testgroupcreate")
	stdOut, stdErr, err := integrationTestWrap([]string{
		"group",
		"create",
		groupname,
		"--displayname",
		groupname,
	})

	assert.NotEmpty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)
	var result lldap.LldapGroup
	err = json.Unmarshal(stdOut.Bytes(), &result)
	assert.Nil(t, err)
	assert.Equal(t, groupname, result.DisplayName)
}

func TestGroupGet(t *testing.T) {

	client := getTestClient()
	testGroup := lldap.LldapGroup{
		DisplayName: "Test Group",
	}
	derr := client.CreateGroup(&testGroup)
	assert.Nil(t, derr)

	groupId := testGroup.Id

	stdOut, stdErr, err := integrationTestWrap([]string{
		"group",
		"get",
		fmt.Sprint(groupId),
	})

	assert.NotEmpty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)
	var result lldap.LldapGroup
	errUnmarshal := json.Unmarshal(stdOut.Bytes(), &result)
	assert.Nil(t, errUnmarshal)
	assert.Equal(t, groupId, result.Id)
	assert.Equal(t, "Test Group", result.DisplayName)
}

func TestGroupUpdate(t *testing.T) {

	client := getTestClient()
	groupname := randomTestSuffix("testgroupupdate")
	testGroup := lldap.LldapGroup{
		DisplayName: groupname,
	}
	derr := client.CreateGroup(&testGroup)
	assert.Nil(t, derr)

	groupId := testGroup.Id
	updatedDisplayName := randomTestSuffix("updatedgroup")
	stdOut, stdErr, err := integrationTestWrap([]string{
		"group",
		"update",
		fmt.Sprint(groupId),
		"--displayname",
		updatedDisplayName,
	})

	assert.Empty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	stdOut, stdErr, err = integrationTestWrap([]string{
		"group",
		"get",
		fmt.Sprint(groupId),
	})

	assert.NotEmpty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)
	var updatedGroup lldap.LldapGroup
	errUnmarshal := json.Unmarshal(stdOut.Bytes(), &updatedGroup)
	assert.Nil(t, errUnmarshal)
	assert.Equal(t, groupId, updatedGroup.Id)
	assert.Equal(t, updatedDisplayName, updatedGroup.DisplayName)
}

func TestGroupDelete(t *testing.T) {

	groupname := randomTestSuffix("testgroupdelete")
	client := getTestClient()
	testGroup := lldap.LldapGroup{
		DisplayName: groupname,
	}
	derr := client.CreateGroup(&testGroup)
	assert.Nil(t, derr)

	groupId := testGroup.Id

	stdOut, stdErr, err := integrationTestWrap([]string{
		"group",
		"delete",
		fmt.Sprint(groupId),
	})

	assert.Empty(t, stdErr)
	assert.Empty(t, stdOut)
	assert.Nil(t, err)

	// Verify group was deleted by trying to get it
	stdOut, stdErr, err = integrationTestWrap([]string{
		"group",
		"get",
		fmt.Sprint(groupId),
	})
	assert.Empty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.NotNil(t, err)
}

func TestMemberAdd(t *testing.T) {

	username := randomTestSuffix("testmemberadduser")
	client := getTestClient()

	testGroup := lldap.LldapGroup{
		DisplayName: "Test Member Add Group",
	}
	derr := client.CreateGroup(&testGroup)
	assert.Nil(t, derr)

	derr = client.CreateUser(&lldap.LldapUser{Id: username, Email: username + "@test.local"})
	assert.Nil(t, derr)

	derr = client.AddUserToGroup(testGroup.Id, username)
	assert.Nil(t, derr)

	stdOut, stdErr, err := integrationTestWrap([]string{
		"group",
		"get",
		fmt.Sprint(testGroup.Id),
	})

	assert.NotEmpty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)
	var group lldap.LldapGroup
	err = json.Unmarshal(stdOut.Bytes(), &group)
	assert.Nil(t, err)

	// Check if user is in the group
	found := false
	for _, member := range group.Users {
		if member.Id == username {
			found = true
			break
		}
	}
	assert.True(t, found, "Member should have been added")
}

func TestAttributeCreate(t *testing.T) {

	attrName := randomTestSuffix("testattributecreate")
	stdOut, stdErr, err := integrationTestWrap([]string{
		"attribute",
		"create",
		attrName,
		"string",
		"--user",
		"--displayname",
		"Test Attribute",
	})

	assert.NotEmpty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)
	var attr lldap.LldapCustomAttribute
	errUnmarshal := json.Unmarshal(stdOut.Bytes(), &attr)
	assert.Nil(t, errUnmarshal)
	assert.Equal(t, attrName, attr.Name)
}

func TestAttributeAdd(t *testing.T) {

	attrName := randomTestSuffix("testattributeadd")
	username := randomTestSuffix("testattributeuser")
	client := getTestClient()

	// Create attribute
	derr := client.CreateUserAttribute(attrName, lldap.LldapCustomAttributeType("STRING"), false, true, true)
	assert.Nil(t, derr)

	derr = client.CreateUser(&lldap.LldapUser{Id: username, Email: username + "@test.local"})
	assert.Nil(t, derr)

	stdOut, stdErr, err := integrationTestWrap([]string{
		"attribute",
		"add",
		attrName,
		username,
		"--user",
		"--values",
		"testvalue",
	})

	assert.Empty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	// Verify attribute was added
	user, derr := client.GetUser(username)
	assert.Nil(t, derr)

	found := false
	for _, attr := range user.Attributes {
		if attr.Name == attrName {
			found = true
			assert.Contains(t, attr.Value, "testvalue")
			break
		}
	}
	assert.True(t, found, "Attribute should have been added to user")
}

func TestAttributeDelete(t *testing.T) {

	attrName := randomTestSuffix("testattributedelete")
	client := getTestClient()

	// Create attribute
	derr := client.CreateUserAttribute(attrName, lldap.LldapCustomAttributeType("STRING"), false, true, true)
	assert.Nil(t, derr)

	stdOut, stdErr, err := integrationTestWrap([]string{
		"attribute",
		"delete",
		attrName,
		"--user",
	})

	assert.Empty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	// Verify attribute was deleted by checking schema
	stdOut, stdErr, err = integrationTestWrap([]string{
		"attribute",
		"schema",
		"--user",
	})

	assert.NotEmpty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	var attrs []lldap.LldapUserAttributeSchema
	err = json.Unmarshal(stdOut.Bytes(), &attrs)
	assert.Nil(t, err)

	found := false
	for _, attr := range attrs {
		if attr.Name == attrName {
			found = true
			break
		}
	}
	assert.False(t, found, "Attribute should have been deleted")
}

func TestAttributeRemove(t *testing.T) {

	attrName := randomTestSuffix("testattributeremove")
	username := randomTestSuffix("testattributeuserremove")
	client := getTestClient()

	// Create attribute
	derr := client.CreateUserAttribute(attrName, lldap.LldapCustomAttributeType("STRING"), false, true, true)
	assert.Nil(t, derr)

	// Create user
	derr = client.CreateUser(&lldap.LldapUser{Id: username, Email: username + "@test.local"})
	assert.Nil(t, derr)

	derr = client.AddAttributeToUser(username, attrName, []string{"testvalue"})
	assert.Nil(t, derr)

	stdOut, stdErr, err := integrationTestWrap([]string{
		"attribute",
		"remove",
		attrName,
		username,
		"--user",
	})

	assert.Empty(t, stdOut)
	assert.Empty(t, stdErr)
	assert.Nil(t, err)

	// Verify attribute was removed from user
	user, derr := client.GetUser(username)
	assert.Nil(t, derr)

	found := false
	for _, attr := range user.Attributes {
		if attr.Name == attrName {
			found = true
			break
		}
	}
	assert.False(t, found, "Attribute should have been removed from user")

	// Clean up
	client.DeleteUser(username)
	client.DeleteUserAttribute(attrName)
}
