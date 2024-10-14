package lldap

import (
	"fmt"
	"net/url"
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

func getTestClient() LldapClient {
	hostIp := os.Getenv("LLDAP_HOST")
	password := os.Getenv("LLDAP_PASSWORD")
	parsedUrl, _ := url.Parse(fmt.Sprintf("http://%s:17170", hostIp))
	client := LldapClient{
		Config: &Config{
			Url:      parsedUrl,
			UserName: "admin",
			Password: password,
		},
	}
	return client
}

func randomTestSuffix(s string) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s-%s", s, string(b))
}

func TestAddUserToGroup(t *testing.T) {
	client := getTestClient()
	groupName := randomTestSuffix("TestAddUserToGroup")
	testGroup := LldapGroup{
		DisplayName: groupName,
	}
	userId := randomTestSuffix("TestAddUserToGroup")
	testUser := LldapUser{
		Id: userId,
	}
	client.CreateGroup(&testGroup)
	client.CreateUser(&testUser)
	result := client.AddUserToGroup(testGroup.Id, testUser.Id)
	assert.Nil(t, result)
	group, _ := client.GetGroup(testGroup.Id)
	users := make([]string, len(group.Users))
	for _, v := range group.Users {
		users = append(users, v.Id)
	}
	assert.True(t, slices.Contains(users, userId))
}

func TestRemoveUserFromGroup(t *testing.T) {
	client := getTestClient()
	groupName := randomTestSuffix("TestRemoveUserFromGroup")
	testGroup := LldapGroup{
		DisplayName: groupName,
	}
	userId := randomTestSuffix("TestRemoveUserFromGroup")
	testUser := LldapUser{
		Id: userId,
	}
	client.CreateGroup(&testGroup)
	client.CreateUser(&testUser)
	client.AddUserToGroup(testGroup.Id, testUser.Id)
	client.RemoveUserFromGroup(testGroup.Id, testUser.Id)
	group, _ := client.GetGroup(testGroup.Id)
	users := make([]string, len(group.Users))
	for _, v := range group.Users {
		users = append(users, v.Id)
	}
	assert.False(t, slices.Contains(users, userId))
}

func TestCreateGroup(t *testing.T) {
	client := getTestClient()
	groupName := randomTestSuffix("TestCreateGroup")
	group := LldapGroup{
		DisplayName: groupName,
	}
	createErr := client.CreateGroup(&group)
	assert.Nil(t, createErr)
	assert.NotEqual(t, 0, group.Id)
	assert.Equal(t, groupName, group.DisplayName)
	assert.NotNil(t, group.DisplayName)
	assert.Equal(t, 0, len(group.Users))
}

func TestUpdateGroupDisplayName(t *testing.T) {
	client := getTestClient()
	initialGroupName := randomTestSuffix("TestUpdateGroupDisplayNameI")
	group := LldapGroup{
		DisplayName: initialGroupName,
	}
	client.CreateGroup(&group)
	assert.NotEqual(t, 0, group.Id)
	expectedGroupName := randomTestSuffix("TestUpdateGroupDisplayNameE")
	updateErr := client.UpdateGroupDisplayName(group.Id, expectedGroupName)
	assert.Nil(t, updateErr)
	result, getErr := client.GetGroup(group.Id)
	assert.Nil(t, getErr)
	assert.Equal(t, expectedGroupName, result.DisplayName)
}

func TestDeleteGroup(t *testing.T) {
	client := getTestClient()
	groupName := randomTestSuffix("TestDeleteGroup")
	testGroup := LldapGroup{
		DisplayName: groupName,
	}
	client.CreateGroup(&testGroup)
	assert.NotEqual(t, 0, testGroup.Id)
	result := client.DeleteGroup(testGroup.Id)
	assert.Nil(t, result)
	groups, _ := client.GetGroups()
	for _, v := range groups {
		assert.False(t, v.DisplayName == groupName)
	}
}

func TestCreateUser(t *testing.T) {
	client := getTestClient()
	userId := randomTestSuffix("TestCreateUser")
	testUser := LldapUser{
		Id: userId,
	}
	client.CreateUser(&testUser)
	result := client.DeleteUser(testUser.Id)
	assert.Nil(t, result)
	users, _ := client.GetUsers()
	for _, v := range users {
		assert.False(t, v.Id == userId)
	}
}

func TestUpdateUser(t *testing.T) {
	client := getTestClient()
	userId := randomTestSuffix("TestUpdateUser")
	testUser := LldapUser{
		Id:          userId,
		Email:       "TestUpdateUser@test.test",
		DisplayName: "Test Update User",
		FirstName:   "Test",
		LastName:    "User",
	}
	client.CreateUser(&testUser)
	testUser.Email = "test@newmail.test"
	testUser.DisplayName = "Real Test User"
	testUser.FirstName = "First"
	testUser.LastName = "Last"
	updateErr := client.UpdateUser(&testUser)
	assert.Nil(t, updateErr)
	user, _ := client.GetUser(userId)
	assert.Equal(t, "test@newmail.test", user.Email)
	assert.Equal(t, "Real Test User", user.DisplayName)
	assert.Equal(t, "First", user.FirstName)
	assert.Equal(t, "Last", user.LastName)
}

func TestDeleteUser(t *testing.T) {
	client := getTestClient()
	userId := randomTestSuffix("TestDeleteUser")
	testUser := LldapUser{
		Id: userId,
	}
	client.CreateUser(&testUser)
	result := client.DeleteUser(testUser.Id)
	assert.Nil(t, result)
	users, _ := client.GetUsers()
	for _, v := range users {
		assert.False(t, v.Id == "TestDeleteUser")
	}
}

func TestGetGroups(t *testing.T) {
	client := getTestClient()
	result, getErr := client.GetGroups()
	assert.Nil(t, getErr)
	assert.NotNil(t, result)
	groupNames := make([]string, len(result))
	for _, v := range result {
		groupNames = append(groupNames, v.DisplayName)
	}
	// LLDAP creates by default:
	// "lldap_admin", "lldap_password_manager", "lldap_strict_readonly"
	assert.True(t, len(result) >= 3)
	assert.True(t, slices.Contains(groupNames, "lldap_admin"))
	assert.True(t, slices.Contains(groupNames, "lldap_password_manager"))
	assert.True(t, slices.Contains(groupNames, "lldap_strict_readonly"))
	for _, group := range result {
		assert.NotEqual(t, "", group.DisplayName)
		assert.NotNil(t, group.CreationDate)
		assert.NotEqual(t, "", group.CreationDate)
	}
}

func TestGetUsers(t *testing.T) {
	client := getTestClient()
	result, getErr := client.GetUsers()
	assert.Nil(t, getErr)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "Administrator", result[0].DisplayName)
}

func TestGetGroup(t *testing.T) {
	client := getTestClient()
	result, getErr := client.GetGroup(1)
	assert.Nil(t, getErr)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Id)
	assert.Equal(t, "lldap_admin", result.DisplayName)
	assert.Equal(t, []LldapUser{
		{
			Id:          "admin",
			DisplayName: "Administrator",
		},
	}, result.Users)
	assert.NotNil(t, result.CreationDate)
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
	assert.NotNil(t, result.CreationDate)
}

func TestGetUserErr(t *testing.T) {
	client := getTestClient()
	result, getErr := client.GetUser("user_does_not_exist")
	assert.NotNil(t, getErr)
	assert.Nil(t, result)
}
