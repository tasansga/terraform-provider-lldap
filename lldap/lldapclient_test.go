package lldap

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

func getTestClient() LldapClient {
	hostIp := os.Getenv("LLDAP_HOST")
	password := os.Getenv("LLDAP_PASSWORD")
	parsedHttpUrl, _ := url.Parse(fmt.Sprintf("http://%s:17170", hostIp))
	parsedLdapUrl, _ := url.Parse(fmt.Sprintf("ldap://%s:3890", hostIp))
	client := LldapClient{
		Config: Config{
			Context:  context.Background(),
			HttpUrl:  parsedHttpUrl,
			LdapUrl:  parsedLdapUrl,
			UserName: "admin",
			Password: password,
			BaseDn:   "dc=tf-provider-lldap,dc=tasansga,dc=github,dc=com",
		},
	}
	return client
}

func randomTestSuffix(s string) string {
	r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	anums := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = anums[r.Intn(len(anums))]
	}
	return fmt.Sprintf("%s-%s", s, string(b))
}

func TestSetUserPassword(t *testing.T) {
	client := getTestClient()
	userId := randomTestSuffix("TestSetUserPassword")
	testUser := LldapUser{
		Id:    userId,
		Email: randomTestSuffix("TestSetUserPasswordEmail"),
	}
	client.CreateUser(&testUser)
	newPassword := randomTestSuffix("TestSetUserPasswordNewPassword")
	result := client.SetUserPassword(strings.ToLower(userId), newPassword)
	assert.Nil(t, result)
	bind, bindErr := client.IsValidPassword(userId, newPassword)
	assert.Nil(t, bindErr)
	assert.NotNil(t, bind)
}

func TestSetUserPasswords(t *testing.T) {
	client := getTestClient()
	// Tests concurrency for ldap:// works as expected
	for i := range 20 {
		userId := randomTestSuffix(fmt.Sprintf("TestSetUserPasswords%d", i))
		testUser := LldapUser{
			Id:    userId,
			Email: randomTestSuffix(fmt.Sprintf("TestSetUserPasswordsEmail%d", i)),
		}
		client.CreateUser(&testUser)
		newPassword := randomTestSuffix(fmt.Sprintf("TestSetUserPasswordsPassword%d", i))
		result := client.SetUserPassword(strings.ToLower(userId), newPassword)
		assert.Nil(t, result)
		bind, bindErr := client.IsValidPassword(userId, newPassword)
		assert.Nil(t, bindErr)
		assert.NotNil(t, bind)
	}
}

func TestAddUserToGroup(t *testing.T) {
	client := getTestClient()
	groupName := randomTestSuffix("TestAddUserToGroup")
	testGroup := LldapGroup{
		DisplayName: groupName,
	}
	userId := randomTestSuffix("TestAddUserToGroup")
	testUser := LldapUser{
		Id:    userId,
		Email: randomTestSuffix("TestAddUserToGroupEmail"),
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
	assert.True(t, slices.Contains(users, strings.ToLower(userId)))
}

func TestRemoveUserFromGroup(t *testing.T) {
	client := getTestClient()
	groupName := randomTestSuffix("TestRemoveUserFromGroup")
	testGroup := LldapGroup{
		DisplayName: groupName,
	}
	userId := randomTestSuffix("TestRemoveUserFromGroup")
	testUser := LldapUser{
		Id:    userId,
		Email: randomTestSuffix("TestRemoveUserFromGroupEmail"),
	}
	client.CreateGroup(&testGroup)
	client.CreateUser(&testUser)
	client.AddUserToGroup(testGroup.Id, testUser.Id)
	response := client.RemoveUserFromGroup(testGroup.Id, testUser.Id)
	assert.Nil(t, response)
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
	assert.NotNil(t, group.Uuid)
	assert.NotEmpty(t, group.Uuid)
	assert.Equal(t, groupName, group.DisplayName)
	assert.NotNil(t, group.DisplayName)
	assert.NotEmpty(t, group.DisplayName)
	assert.Equal(t, 0, len(group.Users))
}

func TestCreateGroups(t *testing.T) {
	client := getTestClient()
	// Tests concurrency for http:// works as expected
	for i := range 20 {
		groupName := randomTestSuffix(fmt.Sprintf("TestCreateGroup%d", i))
		group := LldapGroup{
			DisplayName: groupName,
		}
		createErr := client.CreateGroup(&group)
		assert.Nil(t, createErr)
		assert.NotEqual(t, 0, group.Id)
		assert.NotNil(t, group.Uuid)
		assert.NotEmpty(t, group.Uuid)
		assert.Equal(t, groupName, group.DisplayName)
		assert.NotNil(t, group.DisplayName)
		assert.NotEmpty(t, group.DisplayName)
		assert.Equal(t, 0, len(group.Users))
	}
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
	userEmail := randomTestSuffix("user@email")
	avatar := "/9j/2wBDAAEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQH/wgALCAABAAEBAREA/8QAFAABAAAAAAAAAAAAAAAAAAAAA//aAAgBAQAAAAE//9k="
	testUser := LldapUser{
		Id:          userId,
		Email:       userEmail,
		FirstName:   "fname",
		LastName:    "lname",
		DisplayName: "dname",
		Avatar:      avatar,
	}
	result := client.CreateUser(&testUser)
	assert.NotEmpty(t, testUser.CreationDate)
	assert.NotEmpty(t, testUser.Uuid)
	assert.Nil(t, result)
	assert.Equal(t, userEmail, testUser.Email)
	assert.Equal(t, "fname", testUser.FirstName)
	assert.Equal(t, "lname", testUser.LastName)
	assert.Equal(t, "dname", testUser.DisplayName)
	assert.Equal(t, avatar, testUser.Avatar)
	users, _ := client.GetUsers()
	for _, v := range users {
		assert.False(t, v.Id == userId)
	}
	check, _ := client.GetUser(userId)
	assert.Equal(t, strings.ToLower(userId), check.Id)
	assert.Equal(t, userEmail, check.Email)
	assert.Equal(t, "fname", check.FirstName)
	assert.Equal(t, "lname", check.LastName)
	assert.Equal(t, "dname", check.DisplayName)
	assert.Equal(t, avatar, check.Avatar)
}

func TestUpdateUser(t *testing.T) {
	client := getTestClient()
	userId := randomTestSuffix("TestUpdateUser")
	userEmailI := randomTestSuffix("TestUpdateUser@email.x")
	userEmailE := randomTestSuffix("TestUpdateUser@email.x")
	testUser := LldapUser{
		Id:          userId,
		Email:       userEmailI,
		DisplayName: "Test Update User",
		FirstName:   "Test",
		LastName:    "User",
	}
	client.CreateUser(&testUser)
	testUser.Email = userEmailE
	testUser.DisplayName = "Real Test User"
	testUser.FirstName = "First"
	testUser.LastName = "Last"
	updateErr := client.UpdateUser(&testUser)
	assert.Nil(t, updateErr)
	user, getUserErr := client.GetUser(userId)
	assert.Nil(t, getUserErr)
	assert.Equal(t, userEmailE, user.Email)
	assert.Equal(t, "Real Test User", user.DisplayName)
	assert.Equal(t, "First", user.FirstName)
	assert.Equal(t, "Last", user.LastName)
}

func TestDeleteUser(t *testing.T) {
	client := getTestClient()
	userId := randomTestSuffix("TestDeleteUser")
	email := randomTestSuffix("TestDeleteUser@email.x")
	testUser := LldapUser{
		Id:    userId,
		Email: email,
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
	userNames := make([]string, len(result))
	for _, v := range result {
		userNames = append(userNames, v.Id)
	}
	assert.True(t, slices.Contains(userNames, "admin"))
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
