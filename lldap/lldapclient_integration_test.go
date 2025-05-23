//go:build integration
// +build integration

/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package lldap

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

func getTestClient() LldapClient {
	hostIp := os.Getenv("LLDAP_HOST")
	password := os.Getenv("LLDAP_PASSWORD")
	httpPort := os.Getenv("LLDAP_PORT_HTTP")
	ldapPort := os.Getenv("LLDAP_PORT_LDAP")
	parsedHttpUrl, _ := url.Parse(fmt.Sprintf("http://%s:%s", hostIp, httpPort))
	parsedLdapUrl, _ := url.Parse(fmt.Sprintf("ldap://%s:%s", hostIp, ldapPort))
	client := LldapClient{
		Config: Config{
			Context:  context.Background(),
			HttpUrl:  parsedHttpUrl,
			LdapUrl:  parsedLdapUrl,
			UserName: "admin",
			Password: password,
			BaseDn:   "dc=terraform-provider-lldap,dc=tasansga,dc=github,dc=com",
		},
	}
	return client
}

func randomTestSuffix(s string) string {
	r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	const anums = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = anums[r.Intn(len(anums))]
	}
	return fmt.Sprintf("%s-%s", s, string(b))
}

func TestSetUserPassword(t *testing.T) {
	client := getTestClient()
	userId := randomTestSuffix("TestSetUserPassword")
	createErr := client.CreateUser(&LldapUser{
		Id:    userId,
		Email: userId + "@test.local",
	})
	assert.Nil(t, createErr)

	setErr := client.SetUserPassword(userId, "newpassword")
	assert.Nil(t, setErr)

	// Clean up
	client.DeleteUser(userId)
}

func TestSetUserPasswords(t *testing.T) {
	client := getTestClient()
	// Tests concurrency for ldap:// works as expected
	userIds := []string{
		randomTestSuffix("TestSetUserPasswords1"),
		randomTestSuffix("TestSetUserPasswords2"),
		randomTestSuffix("TestSetUserPasswords3"),
	}

	for _, userId := range userIds {
		createErr := client.CreateUser(&LldapUser{
			Id:    userId,
			Email: userId + "@test.local",
		})
		assert.Nil(t, createErr)
	}

	for _, userId := range userIds {
		setErr := client.SetUserPassword(userId, "newpassword")
		assert.Nil(t, setErr)
	}

	// Clean up
	for _, userId := range userIds {
		client.DeleteUser(userId)
	}
}

func TestGetGroupAttributesSchema(t *testing.T) {
	client := getTestClient()
	getGroupAttr, getGroupAttrErr := client.GetGroupAttributesSchema()
	assert.Nil(t, getGroupAttrErr)
	assert.NotNil(t, getGroupAttr)

	// Check for expected hardcoded attributes
	expectedAttrs := []string{"creation_date", "display_name", "group_id", "uuid"}
	for _, expected := range expectedAttrs {
		found := false
		for _, attr := range getGroupAttr {
			if attr.Name == expected {
				found = true
				assert.True(t, attr.IsHardcoded)
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("Expected hardcoded attribute %s not found", expected))
	}

	// Check that all attributes have required fields
	for _, attr := range getGroupAttr {
		assert.NotEmpty(t, attr.Name)
		assert.NotEmpty(t, attr.AttributeType)
		// IsHardcoded, IsVisible, IsReadonly can be false, so we don't check them
	}

	// Check specific attribute types
	for _, attr := range getGroupAttr {
		switch attr.Name {
		case "creation_date":
			assert.Equal(t, LldapCustomAttributeType("DATE_TIME"), attr.AttributeType)
		case "display_name":
			assert.Equal(t, LldapCustomAttributeType("STRING"), attr.AttributeType)
		case "group_id":
			assert.Equal(t, LldapCustomAttributeType("INTEGER"), attr.AttributeType)
		case "uuid":
			assert.Equal(t, LldapCustomAttributeType("STRING"), attr.AttributeType)
		}
	}
}

func TestGetGroupAttributeSchema(t *testing.T) {
	client := getTestClient()
	attrName := strings.ToLower(randomTestSuffix("TestGetGroupAttributeSchema"))
	createErr := client.CreateGroupAttribute(attrName, LldapCustomAttributeType("STRING"), false, true)
	assert.Nil(t, createErr)

	getGroupAttr, getGroupAttrErr := client.GetGroupAttributeSchema(attrName)
	assert.Nil(t, getGroupAttrErr)
	assert.NotNil(t, getGroupAttr)
	assert.Equal(t, attrName, getGroupAttr.Name)

	// Clean up
	client.DeleteGroupAttribute(attrName)
}

func TestCreateGroupAttribute(t *testing.T) {
	client := getTestClient()
	attrName := strings.ToLower(randomTestSuffix("TestCreateGroupAttribute"))
	createErr := client.CreateGroupAttribute(attrName, LldapCustomAttributeType("STRING"), false, true)
	assert.Nil(t, createErr)

	// Verify it was created
	getGroupAttr, getGroupAttrErr := client.GetGroupAttributeSchema(attrName)
	assert.Nil(t, getGroupAttrErr)
	assert.NotNil(t, getGroupAttr)
	assert.Equal(t, attrName, getGroupAttr.Name)
	assert.Equal(t, LldapCustomAttributeType("STRING"), getGroupAttr.AttributeType)

	// Clean up
	client.DeleteGroupAttribute(attrName)
}

func TestDeleteGroupAttribute(t *testing.T) {
	client := getTestClient()
	attrName := strings.ToLower(randomTestSuffix("TestDeleteGroupAttribute"))
	createErr := client.CreateGroupAttribute(attrName, LldapCustomAttributeType("STRING"), false, true)
	assert.Nil(t, createErr)

	deleteErr := client.DeleteGroupAttribute(attrName)
	assert.Nil(t, deleteErr)

	// Verify it was deleted - check that it's no longer in the list of attributes
	allAttrs, getAllErr := client.GetGroupAttributesSchema()
	assert.Nil(t, getAllErr)
	found := false
	for _, attr := range allAttrs {
		if attr.Name == attrName {
			found = true
			break
		}
	}
	assert.False(t, found, "Deleted attribute should not be found in schema list")
}

func TestGetUserAttributesSchema(t *testing.T) {
	client := getTestClient()
	getUserAttr, getUserAttrErr := client.GetUserAttributesSchema()
	assert.Nil(t, getUserAttrErr)
	assert.NotNil(t, getUserAttr)

	// Check for expected hardcoded attributes
	expectedAttrs := []string{"creation_date", "display_name", "first_name", "last_name", "mail", "user_id", "uuid", "avatar"}
	for _, expected := range expectedAttrs {
		found := false
		for _, attr := range getUserAttr {
			if attr.Name == expected {
				found = true
				assert.True(t, attr.IsHardcoded)
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("Expected hardcoded attribute %s not found", expected))
	}

	// Check that all attributes have required fields
	for _, attr := range getUserAttr {
		assert.NotEmpty(t, attr.Name)
		assert.NotEmpty(t, attr.AttributeType)
		// IsHardcoded, IsVisible, IsReadonly, IsEditable can be false, so we don't check them
	}

	// Check specific attribute types
	for _, attr := range getUserAttr {
		switch attr.Name {
		case "creation_date":
			assert.Equal(t, LldapCustomAttributeType("DATE_TIME"), attr.AttributeType)
		case "display_name", "first_name", "last_name", "mail", "user_id", "uuid":
			assert.Equal(t, LldapCustomAttributeType("STRING"), attr.AttributeType)
		case "avatar":
			assert.Equal(t, LldapCustomAttributeType("JPEG_PHOTO"), attr.AttributeType)
		}
	}

	// Check specific attribute properties
	for _, attr := range getUserAttr {
		switch attr.Name {
		case "creation_date", "user_id", "uuid":
			assert.False(t, attr.IsEditable, fmt.Sprintf("Attribute %s should not be editable", attr.Name))
		case "display_name", "first_name", "last_name", "mail", "avatar":
			assert.True(t, attr.IsEditable, fmt.Sprintf("Attribute %s should be editable", attr.Name))
		}
	}

	// Check readonly attributes
	for _, attr := range getUserAttr {
		switch attr.Name {
		case "creation_date", "user_id", "uuid":
			assert.True(t, attr.IsReadonly, fmt.Sprintf("Attribute %s should be readonly", attr.Name))
		case "display_name", "first_name", "last_name", "mail", "avatar":
			assert.False(t, attr.IsReadonly, fmt.Sprintf("Attribute %s should not be readonly", attr.Name))
		}
	}

	// Check visible attributes (all should be visible)
	for _, attr := range getUserAttr {
		assert.True(t, attr.IsVisible, fmt.Sprintf("Attribute %s should be visible", attr.Name))
	}
}

func TestGetUserAttributeSchema(t *testing.T) {
	client := getTestClient()
	attrName := strings.ToLower(randomTestSuffix("TestGetUserAttributeSchema"))
	createErr := client.CreateUserAttribute(attrName, LldapCustomAttributeType("STRING"), false, true, true)
	assert.Nil(t, createErr)

	getUserAttr, getUserAttrErr := client.GetUserAttributeSchema(attrName)
	assert.Nil(t, getUserAttrErr)
	assert.NotNil(t, getUserAttr)
	assert.Equal(t, attrName, getUserAttr.Name)

	// Clean up
	client.DeleteUserAttribute(attrName)
}

func TestCreateUserAttribute(t *testing.T) {
	client := getTestClient()
	attrName := strings.ToLower(randomTestSuffix("TestCreateUserAttribute"))
	createErr := client.CreateUserAttribute(attrName, LldapCustomAttributeType("STRING"), false, true, true)
	assert.Nil(t, createErr)

	// Verify it was created
	getUserAttr, getUserAttrErr := client.GetUserAttributeSchema(attrName)
	assert.Nil(t, getUserAttrErr)
	assert.NotNil(t, getUserAttr)
	assert.Equal(t, attrName, getUserAttr.Name)
	assert.Equal(t, LldapCustomAttributeType("STRING"), getUserAttr.AttributeType)

	// Clean up
	client.DeleteUserAttribute(attrName)
}

func TestDeleteUserAttribute(t *testing.T) {
	client := getTestClient()
	attrName := strings.ToLower(randomTestSuffix("TestDeleteUserAttribute"))
	createErr := client.CreateUserAttribute(attrName, LldapCustomAttributeType("STRING"), false, true, true)
	assert.Nil(t, createErr)

	deleteErr := client.DeleteUserAttribute(attrName)
	assert.Nil(t, deleteErr)

	// Verify it was deleted - check that it's no longer in the list of attributes
	allAttrs, getAllErr := client.GetUserAttributesSchema()
	assert.Nil(t, getAllErr)
	found := false
	for _, attr := range allAttrs {
		if attr.Name == attrName {
			found = true
			break
		}
	}
	assert.False(t, found, "Deleted attribute should not be found in schema list")
}

func TestAddAttributeToUser(t *testing.T) {
	client := getTestClient()
	attrName := strings.ToLower(randomTestSuffix("TestAddAttributeToUser"))
	userId := randomTestSuffix("TestAddAttributeToUser")

	// Create attribute and user
	createAttrErr := client.CreateUserAttribute(attrName, LldapCustomAttributeType("STRING"), false, true, true)
	assert.Nil(t, createAttrErr)
	createUserErr := client.CreateUser(&LldapUser{Id: userId, Email: userId + "@test.local"})
	assert.Nil(t, createUserErr)

	// Add attribute to user
	addErr := client.AddAttributeToUser(userId, attrName, []string{"test-value"})
	assert.Nil(t, addErr)

	// Verify attribute was added
	user, getUserErr := client.GetUser(userId)
	assert.Nil(t, getUserErr)
	found := false
	for _, attr := range user.Attributes {
		if attr.Name == attrName {
			found = true
			assert.Contains(t, attr.Value, "test-value")
			break
		}
	}
	assert.True(t, found)

	// Clean up
	client.DeleteUser(userId)
	client.DeleteUserAttribute(attrName)
}

func TestRemoveAttributeFromUser(t *testing.T) {
	client := getTestClient()
	attrName := strings.ToLower(randomTestSuffix("TestRemoveAttributeFromUser"))
	userId := randomTestSuffix("TestRemoveAttributeFromUser")

	// Create attribute and user
	createAttrErr := client.CreateUserAttribute(attrName, LldapCustomAttributeType("STRING"), false, true, true)
	assert.Nil(t, createAttrErr)
	createUserErr := client.CreateUser(&LldapUser{Id: userId, Email: userId + "@test.local"})
	assert.Nil(t, createUserErr)

	// Add then remove attribute
	addErr := client.AddAttributeToUser(userId, attrName, []string{"test-value"})
	assert.Nil(t, addErr)
	removeErr := client.RemoveAttributeFromUser(userId, attrName)
	assert.Nil(t, removeErr)

	// Clean up
	client.DeleteUser(userId)
	client.DeleteUserAttribute(attrName)
}

func TestAddUserToGroup(t *testing.T) {
	client := getTestClient()
	groupName := randomTestSuffix("TestAddUserToGroup")
	userId := randomTestSuffix("TestAddUserToGroup")

	// Create group and user
	createGroupErr := client.CreateGroup(&LldapGroup{DisplayName: groupName})
	assert.Nil(t, createGroupErr)
	createUserErr := client.CreateUser(&LldapUser{Id: userId, Email: userId + "@test.local"})
	assert.Nil(t, createUserErr)

	// Get group to find its ID
	groups, getGroupsErr := client.GetGroups()
	assert.Nil(t, getGroupsErr)
	var groupId int
	for _, group := range groups {
		if group.DisplayName == groupName {
			groupId = group.Id
			break
		}
	}

	// Add user to group
	addErr := client.AddUserToGroup(groupId, userId)
	assert.Nil(t, addErr)

	// Clean up
	client.DeleteUser(userId)
	client.DeleteGroup(groupId)
}

func TestRemoveUserFromGroup(t *testing.T) {
	client := getTestClient()
	groupName := randomTestSuffix("TestRemoveUserFromGroup")
	userId := randomTestSuffix("TestRemoveUserFromGroup")

	// Create group and user
	createGroupErr := client.CreateGroup(&LldapGroup{DisplayName: groupName})
	assert.Nil(t, createGroupErr)
	createUserErr := client.CreateUser(&LldapUser{Id: userId, Email: userId + "@test.local"})
	assert.Nil(t, createUserErr)

	// Get group to find its ID
	groups, getGroupsErr := client.GetGroups()
	assert.Nil(t, getGroupsErr)
	var groupId int
	for _, group := range groups {
		if group.DisplayName == groupName {
			groupId = group.Id
			break
		}
	}

	// Add then remove user from group
	addErr := client.AddUserToGroup(groupId, userId)
	assert.Nil(t, addErr)
	removeErr := client.RemoveUserFromGroup(groupId, userId)
	assert.Nil(t, removeErr)

	// Clean up
	client.DeleteUser(userId)
	client.DeleteGroup(groupId)
}

func TestCreateGroup(t *testing.T) {
	client := getTestClient()
	groupName := randomTestSuffix("TestCreateGroup")
	createErr := client.CreateGroup(&LldapGroup{DisplayName: groupName})
	assert.Nil(t, createErr)

	// Verify group was created
	groups, getGroupsErr := client.GetGroups()
	assert.Nil(t, getGroupsErr)
	found := false
	var groupId int
	for _, group := range groups {
		if group.DisplayName == groupName {
			found = true
			groupId = group.Id
			break
		}
	}
	assert.True(t, found)

	// Clean up
	client.DeleteGroup(groupId)
}

func TestCreateGroups(t *testing.T) {
	client := getTestClient()
	// Tests concurrency for http:// works as expected
	groupNames := []string{
		randomTestSuffix("TestCreateGroups1"),
		randomTestSuffix("TestCreateGroups2"),
		randomTestSuffix("TestCreateGroups3"),
	}

	for _, groupName := range groupNames {
		createErr := client.CreateGroup(&LldapGroup{DisplayName: groupName})
		assert.Nil(t, createErr)
	}

	// Verify groups were created
	groups, getGroupsErr := client.GetGroups()
	assert.Nil(t, getGroupsErr)

	// Clean up
	for _, groupName := range groupNames {
		for _, group := range groups {
			if group.DisplayName == groupName {
				client.DeleteGroup(group.Id)
				break
			}
		}
	}
}

func TestUpdateGroupDisplayName(t *testing.T) {
	client := getTestClient()
	initialGroupName := randomTestSuffix("TestUpdateGroupDisplayNameI")
	updatedGroupName := randomTestSuffix("TestUpdateGroupDisplayNameU")
	createErr := client.CreateGroup(&LldapGroup{DisplayName: initialGroupName})
	assert.Nil(t, createErr)

	// Get group ID
	groups, getGroupsErr := client.GetGroups()
	assert.Nil(t, getGroupsErr)
	var groupId int
	for _, group := range groups {
		if group.DisplayName == initialGroupName {
			groupId = group.Id
			break
		}
	}

	// Update group display name
	updateErr := client.UpdateGroupDisplayName(groupId, updatedGroupName)
	assert.Nil(t, updateErr)

	// Clean up
	client.DeleteGroup(groupId)
}

func TestDeleteGroup(t *testing.T) {
	client := getTestClient()
	groupName := randomTestSuffix("TestDeleteGroup")
	createErr := client.CreateGroup(&LldapGroup{DisplayName: groupName})
	assert.Nil(t, createErr)

	// Get group ID
	groups, getGroupsErr := client.GetGroups()
	assert.Nil(t, getGroupsErr)
	var groupId int
	for _, group := range groups {
		if group.DisplayName == groupName {
			groupId = group.Id
			break
		}
	}

	// Delete group
	deleteErr := client.DeleteGroup(groupId)
	assert.Nil(t, deleteErr)

	// Verify group was deleted
	_, getGroupErr := client.GetGroup(groupId)
	assert.NotNil(t, getGroupErr)
}

func TestCreateUser(t *testing.T) {
	client := getTestClient()
	userId := strings.ToLower(randomTestSuffix("TestCreateUser")) // Use lowercase to match LLDAP behavior
	createErr := client.CreateUser(&LldapUser{
		Id:          userId,
		Email:       userId + "@test.local",
		DisplayName: "Test User",
		FirstName:   "Test",
		LastName:    "User",
	})
	assert.Nil(t, createErr)

	// Verify user was created
	user, getUserErr := client.GetUser(userId)
	assert.Nil(t, getUserErr)
	assert.Equal(t, userId, user.Id)
	assert.Equal(t, userId+"@test.local", user.Email)
	assert.Equal(t, "Test User", user.DisplayName)
	assert.Equal(t, "Test", user.FirstName)
	assert.Equal(t, "User", user.LastName)

	// Check that user has expected attributes
	expectedAttrs := []string{"creation_date", "display_name", "first_name", "last_name", "mail", "user_id", "uuid", "avatar"}
	for _, expected := range expectedAttrs {
		found := false
		for _, attr := range user.Attributes {
			if attr.Name == expected {
				found = true
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("Expected attribute %s not found", expected))
	}

	// Clean up
	client.DeleteUser(userId)
}

func TestUpdateUser(t *testing.T) {
	client := getTestClient()
	userId := randomTestSuffix("TestUpdateUser")
	createErr := client.CreateUser(&LldapUser{
		Id:    userId,
		Email: userId + "@test.local",
	})
	assert.Nil(t, createErr)

	// Update user
	updateErr := client.UpdateUser(&LldapUser{
		Id:          userId,
		Email:       userId + "@updated.local",
		DisplayName: "Updated User",
		FirstName:   "Updated",
		LastName:    "User",
	})
	assert.Nil(t, updateErr)

	// Verify user was updated
	user, getUserErr := client.GetUser(userId)
	assert.Nil(t, getUserErr)
	assert.Equal(t, userId+"@updated.local", user.Email)
	assert.Equal(t, "Updated User", user.DisplayName)
	assert.Equal(t, "Updated", user.FirstName)
	assert.Equal(t, "User", user.LastName)

	// Clean up
	client.DeleteUser(userId)
}

func TestDeleteUser(t *testing.T) {
	client := getTestClient()
	userId := randomTestSuffix("TestDeleteUser")
	createErr := client.CreateUser(&LldapUser{
		Id:    userId,
		Email: userId + "@test.local",
	})
	assert.Nil(t, createErr)

	// Delete user
	deleteErr := client.DeleteUser(userId)
	assert.Nil(t, deleteErr)

	// Verify user was deleted
	_, getUserErr := client.GetUser(userId)
	assert.NotNil(t, getUserErr)
}

func TestGetGroups(t *testing.T) {
	client := getTestClient()
	result, getErr := client.GetGroups()
	assert.Nil(t, getErr)
	assert.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Check that default groups exist
	defaultGroups := []string{"lldap_admin", "lldap_password_manager", "lldap_strict_readonly"}
	for _, defaultGroup := range defaultGroups {
		found := false
		for _, group := range result {
			if group.DisplayName == defaultGroup {
				found = true
				assert.Greater(t, group.Id, 0)
				assert.NotEmpty(t, group.CreationDate)
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("Default group %s not found", defaultGroup))
	}
}

func TestGetUsers(t *testing.T) {
	client := getTestClient()
	result, getErr := client.GetUsers()
	assert.Nil(t, getErr)
	assert.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Check that admin user exists
	found := false
	for _, user := range result {
		if user.Id == "admin" {
			found = true
			assert.Equal(t, "Administrator", user.DisplayName)
			break
		}
	}
	assert.True(t, found, "Admin user not found")
}

func TestGetGroup(t *testing.T) {
	client := getTestClient()
	result, getErr := client.GetGroup(1)
	assert.Nil(t, getErr)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Id)
	assert.Equal(t, "lldap_admin", result.DisplayName)
	assert.NotEmpty(t, result.CreationDate)

	// Check that group has expected attributes
	expectedAttrs := []string{"creation_date", "display_name", "group_id", "uuid"}
	for _, expected := range expectedAttrs {
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

func TestGetGroupErr(t *testing.T) {
	client := getTestClient()
	result, getErr := client.GetGroup(-2)
	assert.NotNil(t, getErr)
	assert.Nil(t, result)
}

func TestGetUser(t *testing.T) {
	client := getTestClient()
	result, getErr := client.GetUser("admin")
	assert.Nil(t, getErr)
	assert.NotNil(t, result)
	assert.Equal(t, "admin", result.Id)
	assert.Equal(t, "Administrator", result.DisplayName)

	// Check that user has core required attributes (some may be empty but should exist)
	requiredAttrs := []string{"creation_date", "display_name", "mail", "user_id", "uuid"}
	for _, required := range requiredAttrs {
		found := false
		for _, attr := range result.Attributes {
			if attr.Name == required {
				found = true
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("Required attribute %s not found", required))
	}

	// Check that attributes have proper structure
	for _, attr := range result.Attributes {
		assert.NotEmpty(t, attr.Name, "Attribute name should not be empty")
		assert.NotNil(t, attr.Value, "Attribute value should not be nil")
	}
}

func TestGetUserErr(t *testing.T) {
	client := getTestClient()
	result, getErr := client.GetUser("user_does_not_exist")
	assert.NotNil(t, getErr)
	assert.Nil(t, result)
}
