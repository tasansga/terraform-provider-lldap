package lldap

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestClient() LldapClient {
	hostIp := os.Getenv("LLDAP_HOST")
	password := os.Getenv("LLDAP_PASSWORD")
	client := LldapClient{
		Config: &Config{
			Url:      fmt.Sprintf("http://%s:17170", hostIp),
			UserName: "admin",
			Password: password,
		},
	}
	return client
}

func TestAddUserToGroup(t *testing.T) {
	// TODO
}

func TestRemoveUserFromGroup(t *testing.T) {
	// TODO
}

func TestCreateGroup(t *testing.T) {
	// TODO
}

func TestUpdateGroup(t *testing.T) {
	// TODO
}

func TestDeleteGroup(t *testing.T) {
	// TODO
}

func TestCreateUser(t *testing.T) {
	// TODO
}

func TestUpdateUser(t *testing.T) {
	// TODO
}

func TestDeleteUser(t *testing.T) {
	// TODO
}

func TestGetGroups(t *testing.T) {
	client := getTestClient()
	result, getErr := client.GetGroups()
	assert.Nil(t, getErr)
	assert.NotNil(t, result)
	// LLDAP creates by default:
	// "lldap_admin", "lldap_password_manager", "lldap_strict_readonly"
	assert.Equal(t, 3, len(result))
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
