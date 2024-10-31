package lldap

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

/*
LLDAP GraphQL "HowTo"

GraphQL schema:
https://github.com/lldap/lldap/blob/main/schema.graphql

Pre-defined GraphQL queries:
https://github.com/lldap/lldap/tree/main/app/queries

How to use the GraphQL API with `curl`:
https://github.com/lldap/lldap/blob/main/scripts/bootstrap.sh

	Authentication
	```
		response="$(curl --silent --request POST \
		--url "$url/auth/simple/login" \
		--header 'Content-Type: application/json' \
		--data "$(jo -- username="$admin_username" password="$admin_password")")"
	```

	Query file
	```
		curl --silent --request POST \
		--url "$LLDAP_URL/api/graphql" \
		--header "Authorization: Bearer $TOKEN" \
		--header 'Content-Type: application/json' \
		--data @<(jq --slurpfile variables "$variables_file" '. + {"variables": $variables[0]}' "$query_file")
	```
*/

type LldapClientQuery struct {
	Query         string      `json:"query"`
	OperationName string      `json:"operationName"`
	Variables     interface{} `json:"variables"`
}

type LldapClientError struct {
	Message   string        `json:"message"`
	Locations []interface{} `json:"locations"`
	Path      []string      `json:"path"`
}

type LldapClientResponse[T interface{}] struct {
	Data   *T                 `json:"data"`
	Errors []LldapClientError `json:"errors"`
}

type LldapMutateOk struct {
	OK bool `json:"ok"`
}

type LldapGroup struct {
	Id           int         `json:"id"`
	DisplayName  string      `json:"displayName"`
	CreationDate string      `json:"creationDate"`
	Uuid         string      `json:"uuid"`
	Users        []LldapUser `json:"users"`
}

type LldapUser struct {
	Id           string       `json:"id"`
	Password     string       `json:"password"`
	Email        string       `json:"email"`
	DisplayName  string       `json:"displayName"`
	FirstName    string       `json:"firstName"`
	LastName     string       `json:"lastName"`
	CreationDate string       `json:"creationDate"`
	Uuid         string       `json:"uuid"`
	Avatar       string       `json:"avatar"`
	Groups       []LldapGroup `json:"groups"`
}

type LldapClient struct {
	Config       Config
	Token        string
	RefreshToken string
	HttpClient   *http.Client
	LdapClient   *ldap.Conn
}

func getLdapBindConnection(ldapUrl string, baseDn string, username string, password string) (*ldap.Conn, diag.Diagnostics) {
	ldapclient, dialErr := ldap.DialURL(ldapUrl, ldap.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}))
	if dialErr != nil {
		return nil, diag.Errorf("unable to dial ldap url: %s", dialErr)
	}
	userDn := fmt.Sprintf("cn=%s,ou=people,%s", ldap.EscapeFilter(username), baseDn)
	bindErr := ldapclient.Bind(userDn, password)
	if bindErr != nil {
		return nil, diag.Errorf("could not bind to ldap server: %s", bindErr)
	}
	return ldapclient, nil
}

func (lc *LldapClient) IsValidPassword(username string, password string) (bool, diag.Diagnostics) {
	bind, bindErr := getLdapBindConnection(lc.Config.LdapUrl.String(), lc.Config.BaseDn, username, password)
	if bindErr != nil {
		if strings.Contains(bindErr[len(bindErr)-1].Summary, "Invalid Credentials") {
			return false, nil
		} else {
			return false, bindErr
		}
	}
	defer bind.Close()
	return true, nil
}

func (lc *LldapClient) SetUserPassword(username string, newPassword string) diag.Diagnostics {
	if lc.LdapClient == nil {
		ldapclient, bindErr := getLdapBindConnection(lc.Config.LdapUrl.String(), lc.Config.BaseDn, lc.Config.UserName, lc.Config.Password)
		if bindErr != nil {
			return bindErr
		}
		lc.LdapClient = ldapclient
	}
	userDn := fmt.Sprintf("cn=%s,ou=people,%s", ldap.EscapeFilter(username), lc.Config.BaseDn)
	_, modifyErr := lc.LdapClient.PasswordModify(&ldap.PasswordModifyRequest{
		UserIdentity: userDn,
		NewPassword:  newPassword,
	})
	if modifyErr != nil {
		return diag.Errorf("unable to modify password for '%s': %s", userDn, modifyErr)
	}
	return nil
}

func (lc *LldapClient) query(query LldapClientQuery) ([]byte, diag.Diagnostics) {
	if lc.Token == "" {
		authErr := lc.Authenticate()
		if authErr != nil {
			return nil, authErr
		}
	}
	queryJson, marshErr := json.Marshal(query)
	if marshErr != nil {
		return nil, diag.FromErr(marshErr)
	}
	ref, _ := url.Parse("/api/graphql")
	graphQlApiUrl := lc.Config.HttpUrl.ResolveReference(ref)
	req, reqErr := http.NewRequest("POST", graphQlApiUrl.String(), strings.NewReader(string(queryJson)))
	if reqErr != nil {
		return nil, diag.FromErr(reqErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", lc.Token))
	resp, respErr := lc.HttpClient.Do(req)
	if respErr != nil {
		return nil, diag.FromErr(respErr)
	}
	defer resp.Body.Close()
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, diag.FromErr(readErr)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, diag.Errorf("Unexpected HTTP status code in response: %d - %s", resp.StatusCode, string(bodyBytes))
	}
	return bodyBytes, nil
}

func (lc *LldapClient) Authenticate() diag.Diagnostics {
	if lc.HttpClient == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: lc.Config.InsecureSkipCertCheck},
		}
		lc.HttpClient = &http.Client{Transport: tr}
	}
	type AuthBody struct {
		UserName string `json:"username"`
		Password string `json:"password"`
	}
	type AuthResponse struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
	}
	authBody, marshErr := json.Marshal(&AuthBody{
		UserName: lc.Config.UserName,
		Password: lc.Config.Password,
	})
	if marshErr != nil {
		return diag.FromErr(marshErr)
	}
	ref, _ := url.Parse("/auth/simple/login")
	authSimpleUrl := lc.Config.HttpUrl.ResolveReference(ref)
	req, reqErr := http.NewRequest("POST", authSimpleUrl.String(), strings.NewReader(string(authBody)))
	if reqErr != nil {
		return diag.FromErr(reqErr)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, respErr := lc.HttpClient.Do(req)
	if respErr != nil {
		return diag.FromErr(respErr)
	}
	defer resp.Body.Close()
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return diag.FromErr(readErr)
	}
	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("Unexpected HTTP status code in response: %d - %s", resp.StatusCode, string(bodyBytes))
	}
	authResponse := AuthResponse{}
	unmarshErr := json.Unmarshal(bodyBytes, &authResponse)
	if unmarshErr != nil {
		return diag.FromErr(unmarshErr)
	}
	lc.Token = authResponse.Token
	lc.RefreshToken = authResponse.RefreshToken
	return nil
}

func (lc *LldapClient) AddUserToGroup(groupId int, userId string) diag.Diagnostics {
	type AddUserToGroupVariables struct {
		UserId  string `json:"user"`
		GroupId int    `json:"group"`
	}
	type AddUserResponseData struct {
		AddUserToGroup LldapMutateOk `json:"addUserToGroup"`
	}
	query := LldapClientQuery{
		Query:         "mutation AddUserToGroup($user: String!, $group: Int!) {addUserToGroup(userId: $user, groupId: $group) {ok}}",
		OperationName: "AddUserToGroup",
		Variables: AddUserToGroupVariables{
			UserId:  userId,
			GroupId: groupId,
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return responseDiagErr
	}
	addUserResponse := LldapClientResponse[AddUserResponseData]{}
	unmarshErr := json.Unmarshal(response, &addUserResponse)
	if unmarshErr != nil {
		return diag.Errorf("Could not unmarshal response: %s", unmarshErr)
	}
	if addUserResponse.Errors != nil {
		return diag.Errorf("GraphQL query returned error: %s", string(response))
	}
	if !addUserResponse.Data.AddUserToGroup.OK {
		return diag.Errorf("Failed to add user to group: %s", string(response))
	}
	return nil
}

func (lc *LldapClient) RemoveUserFromGroup(groupId int, userId string) diag.Diagnostics {
	type RemoveUserFromGroupVariables struct {
		UserId  string `json:"user"`
		GroupId int    `json:"group"`
	}
	type RemoveUserResponseData struct {
		RemoveUserFromGroup LldapMutateOk `json:"removeUserFromGroup"`
	}
	query := LldapClientQuery{
		Query:         "mutation RemoveUserFromGroup($user: String!, $group: Int!) {removeUserFromGroup(userId: $user, groupId: $group) {ok}}",
		OperationName: "RemoveUserFromGroup",
		Variables: RemoveUserFromGroupVariables{
			UserId:  userId,
			GroupId: groupId,
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return responseDiagErr
	}
	removeUserResponse := LldapClientResponse[RemoveUserResponseData]{}
	unmarshErr := json.Unmarshal(response, &removeUserResponse)
	if unmarshErr != nil {
		return diag.Errorf("Could not unmarshal response: %s", unmarshErr)
	}
	if removeUserResponse.Errors != nil {
		return diag.Errorf("GraphQL query returned error: %s", string(response))
	}
	if !removeUserResponse.Data.RemoveUserFromGroup.OK {
		return diag.Errorf("Failed to add user to group: %s", string(response))
	}
	return nil
}

func (lc *LldapClient) CreateGroup(group *LldapGroup) diag.Diagnostics {
	type CreateGroupVariables struct {
		Name string `json:"name"`
	}
	type GroupResponseData struct {
		Group LldapGroup `json:"createGroup"`
	}
	query := LldapClientQuery{
		Query:         "mutation CreateGroup($name: String!) {createGroup(name: $name) {id displayName uuid}}",
		OperationName: "CreateGroup",
		Variables: CreateGroupVariables{
			Name: group.DisplayName,
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return responseDiagErr
	}
	newGroupResponse := LldapClientResponse[GroupResponseData]{}
	unmarshErr := json.Unmarshal(response, &newGroupResponse)
	if unmarshErr != nil {
		return diag.Errorf("Could not unmarshal response: %s", unmarshErr)
	}
	if newGroupResponse.Errors != nil {
		return diag.Errorf("GraphQL query returned error: %s", string(response))
	}
	for _, user := range group.Users {
		addUserErr := lc.AddUserToGroup(newGroupResponse.Data.Group.Id, user.Id)
		if addUserErr != nil {
			return addUserErr
		}
	}
	getGroup, getGroupErr := lc.GetGroup(newGroupResponse.Data.Group.Id)
	if getGroupErr != nil {
		return getGroupErr
	}
	group.Id = newGroupResponse.Data.Group.Id
	group.CreationDate = getGroup.CreationDate
	group.DisplayName = getGroup.DisplayName
	group.Uuid = getGroup.Uuid
	group.Users = getGroup.Users
	return nil
}

func (lc *LldapClient) GetGroup(id int) (*LldapGroup, diag.Diagnostics) {
	type GetGroupVariables struct {
		Id int `json:"id"`
	}
	type LldapGroupResponseData struct {
		Group LldapGroup `json:"group"`
	}
	query := LldapClientQuery{
		Query:         "query GetGroupDetails($id: Int!) {group(groupId: $id) {id displayName creationDate uuid users {id displayName}}}",
		OperationName: "GetGroupDetails",
		Variables: GetGroupVariables{
			Id: id,
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return nil, responseDiagErr
	}
	group := LldapClientResponse[LldapGroupResponseData]{}
	unmarshErr := json.Unmarshal(response, &group)
	if unmarshErr != nil {
		return nil, diag.FromErr(unmarshErr)
	}
	if group.Errors != nil {
		return nil, diag.Errorf("GraphQL query returned error: %s", string(response))
	}
	return &group.Data.Group, nil
}

func (lc *LldapClient) UpdateGroupDisplayName(groupId int, displayName string) diag.Diagnostics {
	type UpdateGroupInput struct {
		Id          int    `json:"id"`
		DisplayName string `json:"displayName"`
	}
	type UpdateGroupDisplayNameVariables struct {
		UpdateGroup UpdateGroupInput `json:"group"`
	}
	query := LldapClientQuery{
		Query:         "mutation UpdateGroup($group: UpdateGroupInput!) {updateGroup(group: $group) {ok}}",
		OperationName: "UpdateGroup",
		Variables: UpdateGroupDisplayNameVariables{
			UpdateGroup: UpdateGroupInput{
				Id:          groupId,
				DisplayName: displayName,
			},
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return responseDiagErr
	}
	type LldapUpdateGroupResponseData struct {
		UpdateGroup LldapMutateOk `json:"updateGroup"`
	}
	updateResponse := LldapClientResponse[LldapUpdateGroupResponseData]{}
	unmarshErr := json.Unmarshal(response, &updateResponse)
	if unmarshErr != nil {
		return diag.FromErr(unmarshErr)
	}
	if updateResponse.Errors != nil {
		return diag.Errorf("GraphQL query returned error: %s", string(response))
	}
	if !updateResponse.Data.UpdateGroup.OK {
		return diag.Errorf("Failed to update group display name: %s", string(response))
	}
	return nil
}

func (lc *LldapClient) DeleteGroup(id int) diag.Diagnostics {
	type DeleteGroupVariables struct {
		GroupId int `json:"groupId"`
	}
	query := LldapClientQuery{
		Query:         "mutation DeleteGroupQuery($groupId: Int!) {deleteGroup(groupId: $groupId) {ok}}",
		OperationName: "DeleteGroupQuery",
		Variables: DeleteGroupVariables{
			GroupId: id,
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return responseDiagErr
	}
	type LldapDeleteGroupResponseData struct {
		DeleteGroup LldapMutateOk `json:"deleteGroup"`
	}
	deleteResponse := LldapClientResponse[LldapDeleteGroupResponseData]{}
	unmarshErr := json.Unmarshal(response, &deleteResponse)
	if unmarshErr != nil {
		return diag.FromErr(unmarshErr)
	}
	if deleteResponse.Errors != nil {
		return diag.Errorf("GraphQL query returned error: %s", string(response))
	}
	if !deleteResponse.Data.DeleteGroup.OK {
		return diag.Errorf("Failed to delete group: %s", string(response))
	}
	return nil
}

func (lc *LldapClient) CreateUser(user *LldapUser) diag.Diagnostics {
	type CreateUserInput struct {
		Id          string `json:"id"`
		DisplayName string `json:"displayName"`
		Email       string `json:"email"`
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Avatar      string `json:"avatar"`
	}
	type CreateUserVariables struct {
		CreateUserInput CreateUserInput `json:"user"`
	}
	query := LldapClientQuery{
		Query:         "mutation CreateUser($user: CreateUserInput!) {createUser(user: $user) {id creationDate uuid avatar}}",
		OperationName: "CreateUser",
		Variables: CreateUserVariables{
			CreateUserInput: CreateUserInput{
				Id:          user.Id,
				DisplayName: user.DisplayName,
				Email:       user.Email,
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				Avatar:      user.Avatar,
			},
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return responseDiagErr
	}
	type LldapCreateUserResponseData struct {
		Id           string `json:"id"`
		CreationDate string `json:"creationDate"`
		Uuid         string `json:"uuid"`
	}
	type LldapCreateUserResponse struct {
		CreateUser LldapCreateUserResponseData `json:"createUser"`
	}
	createdUser := LldapClientResponse[LldapCreateUserResponse]{}
	unmarshErr := json.Unmarshal(response, &createdUser)
	if unmarshErr != nil {
		return diag.FromErr(unmarshErr)
	}
	if createdUser.Errors != nil {
		return diag.Errorf("GraphQL query returned error: %s (%s)", string(response), user.Id)
	}
	user.CreationDate = createdUser.Data.CreateUser.CreationDate
	user.Uuid = createdUser.Data.CreateUser.Uuid
	return nil
}

func (lc *LldapClient) GetUser(id string) (*LldapUser, diag.Diagnostics) {
	type GetUserVariables struct {
		Id string `json:"id"`
	}
	type LldapUserResponseData struct {
		User LldapUser `json:"user"`
	}
	query := LldapClientQuery{
		Query:         "query GetUserDetails($id: String!) {user(userId: $id) {id email displayName firstName lastName creationDate uuid avatar groups {id displayName}}}",
		OperationName: "GetUserDetails",
		Variables: GetUserVariables{
			Id: id,
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return nil, responseDiagErr
	}
	user := LldapClientResponse[LldapUserResponseData]{}
	unmarshErr := json.Unmarshal(response, &user)
	if unmarshErr != nil {
		return nil, diag.FromErr(unmarshErr)
	}
	if user.Errors != nil {
		return nil, diag.Errorf("GraphQL query returned error: %s", string(response))
	}
	return &user.Data.User, nil
}

func (lc *LldapClient) UpdateUser(user *LldapUser) diag.Diagnostics {
	type UpdateUserInput struct {
		Id          string `json:"id"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Avatar      string `json:"avatar"`
	}
	type UpdateUserVariable struct {
		UpdateUser UpdateUserInput `json:"user"`
	}
	query := LldapClientQuery{
		Query:         "mutation UpdateUser($user: UpdateUserInput!) {updateUser(user: $user) {ok}}",
		OperationName: "UpdateUser",
		Variables: UpdateUserVariable{
			UpdateUser: UpdateUserInput{
				Id:          user.Id,
				Email:       user.Email,
				DisplayName: user.DisplayName,
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				Avatar:      user.Avatar,
			},
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return responseDiagErr
	}
	type LldapUpdateUserResponseData struct {
		UpdateUser LldapMutateOk `json:"updateUser"`
	}
	updateResponse := LldapClientResponse[LldapUpdateUserResponseData]{}
	unmarshErr := json.Unmarshal(response, &updateResponse)
	if unmarshErr != nil {
		return diag.FromErr(unmarshErr)
	}
	if updateResponse.Errors != nil {
		return diag.Errorf("GraphQL query returned error: %s", string(response))
	}
	if !updateResponse.Data.UpdateUser.OK {
		return diag.Errorf("Failed to update user: %s", string(response))
	}
	return nil
}

func (lc *LldapClient) DeleteUser(id string) diag.Diagnostics {
	type DeleteUserVariable struct {
		Id string `json:"user"`
	}
	query := LldapClientQuery{
		Query:         "mutation DeleteUserQuery($user: String!) {deleteUser(userId: $user) {ok}}",
		OperationName: "DeleteUserQuery",
		Variables: DeleteUserVariable{
			Id: id,
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return responseDiagErr
	}
	type LldapDeleteUserResponseData struct {
		DeleteUser LldapMutateOk `json:"deleteUser"`
	}
	deleteResponse := LldapClientResponse[LldapDeleteUserResponseData]{}
	unmarshErr := json.Unmarshal(response, &deleteResponse)
	if unmarshErr != nil {
		return diag.FromErr(unmarshErr)
	}
	if deleteResponse.Errors != nil {
		return diag.Errorf("GraphQL query returned error: %s", string(response))
	}
	if !deleteResponse.Data.DeleteUser.OK {
		return diag.Errorf("Failed to delete user: %s", string(response))
	}
	return nil
}

func (lc *LldapClient) GetGroups() ([]LldapGroup, diag.Diagnostics) {
	type LldapGroupListResponseData struct {
		Groups []LldapGroup `json:"groups"`
	}
	query := LldapClientQuery{
		Query:         "query GetGroupList {groups {id displayName creationDate}}",
		OperationName: "GetGroupList",
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return nil, responseDiagErr
	}
	groups := LldapClientResponse[LldapGroupListResponseData]{}
	unmarshErr := json.Unmarshal(response, &groups)
	if unmarshErr != nil {
		return nil, diag.FromErr(unmarshErr)
	}
	if groups.Errors != nil {
		return nil, diag.Errorf("GraphQL query returned error: %s", string(response))
	}
	return groups.Data.Groups, nil
}

func (lc *LldapClient) GetUsers() ([]LldapUser, diag.Diagnostics) {
	type LldapUserListResponseData struct {
		Users []LldapUser `json:"users"`
	}
	query := LldapClientQuery{
		Query:         "query ListUsersQuery($filters: RequestFilter) {users(filters: $filters) {id email displayName firstName lastName creationDate uuid avatar}}",
		OperationName: "ListUsersQuery",
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return nil, responseDiagErr
	}
	users := LldapClientResponse[LldapUserListResponseData]{}
	unmarshErr := json.Unmarshal(response, &users)
	if unmarshErr != nil {
		return nil, diag.FromErr(unmarshErr)
	}
	if users.Errors != nil {
		return nil, diag.Errorf("GraphQL query returned error: %s", string(response))
	}
	return users.Data.Users, nil
}
