package lldap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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

type LldapGroup struct {
	Id           int         `json:"id"`
	DisplayName  string      `json:"displayName"`
	CreationDate *string     `json:"creationDate"`
	Users        []LldapUser `json:"users"`
}

type LldapUser struct {
	Id           string       `json:"id"`
	Email        string       `json:"email"`
	DisplayName  string       `json:"displayName"`
	FirstName    string       `json:"firstName"`
	LastName     string       `json:"lastName"`
	CreationDate string       `json:"creationDate"`
	Groups       []LldapGroup `json:"groups"`
}

type LldapClient struct {
	Config       *Config
	Token        string
	RefreshToken string
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
	graphQlApiUrl := lc.Config.Url.ResolveReference(ref)
	req, reqErr := http.NewRequest("POST", graphQlApiUrl.String(), strings.NewReader(string(queryJson)))
	if reqErr != nil {
		return nil, diag.FromErr(reqErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", lc.Token))
	resp, respErr := http.DefaultClient.Do(req)
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
	authSimpleUrl := lc.Config.Url.ResolveReference(ref)
	req, reqErr := http.NewRequest("POST", authSimpleUrl.String(), strings.NewReader(string(authBody)))
	if reqErr != nil {
		return diag.FromErr(reqErr)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, respErr := http.DefaultClient.Do(req)
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
	// {"query":"mutation AddUserToGroup($user: String!, $group: Int!) {addUserToGroup(userId: $user, groupId: $group) {ok}}","operationName":"AddUserToGroup"}
	// TODO
	return diag.Errorf("Not implemented")
}

func (lc *LldapClient) RemoveUserFromGroup(groupId int, userId string) diag.Diagnostics {
	// {"operationName":"RemoveUserFromGroup","query":"mutation RemoveUserFromGroup($user: String!, $group: Int!) {removeUserFromGroup(userId: $user, groupId: $group) {ok}}"}
	// TODO
	return diag.Errorf("Not implemented")
}

func (lc *LldapClient) CreateGroup(group *LldapGroup) diag.Diagnostics {
	type CreateGroupVariables struct {
		Name string `json:"name"`
	}
	type LldapGroupResponseData struct {
		Group LldapGroup `json:"createGroup"`
	}
	query := LldapClientQuery{
		Query:         "mutation CreateGroup($name: String!) {createGroup(name: $name) {id displayName}}",
		OperationName: "CreateGroup",
		Variables: CreateGroupVariables{
			Name: group.DisplayName,
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return responseDiagErr
	}
	newGroupResponse := LldapClientResponse[LldapGroupResponseData]{}
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
		Query:         "query GetGroupDetails($id: Int!) {group(groupId: $id) {id displayName creationDate users {id displayName}}}",
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
	type LldapMutateOk struct {
		OK bool `json:"ok"`
	}
	type LldapUpdateGroupResponseData struct {
		UpdateGroup LldapMutateOk `json:"updateGroup"`
	}
	updateResponse := LldapClientResponse[LldapUpdateGroupResponseData]{}
	unmarshErr := json.Unmarshal(response, &updateResponse)
	if unmarshErr != nil {
		return diag.FromErr(unmarshErr)
	}
	if !updateResponse.Data.UpdateGroup.OK {
		return diag.Errorf("Failed to update group display name: %s", string(response))
	}
	return nil
}

func (lc *LldapClient) DeleteGroup(id int) diag.Diagnostics {
	type DeleteGroupVariables struct {
		Id int `json:"id"`
	}
	query := LldapClientQuery{
		Query:         "mutation DeleteGroupQuery($groupId: Int!) {deleteGroup(groupId: $groupId) {ok}}",
		OperationName: "DeleteGroupQuery",
		Variables: DeleteGroupVariables{
			Id: id,
		},
	}
	response, responseDiagErr := lc.query(query)
	if responseDiagErr != nil {
		return responseDiagErr
	}
	// TODO
	return diag.Errorf("Not implemented: DeleteGroup %s", string(response))
}

func (lc *LldapClient) CreateUser(user *LldapUser) diag.Diagnostics {
	// {"query":"mutation CreateUser($user: CreateUserInput!) {createUser(user: $user) {id creationDate}}","operationName":"CreateUser"}
	// TODO
	return diag.Errorf("Not implemented: CreateUser")
}

func (lc *LldapClient) GetUser(id string) (*LldapUser, diag.Diagnostics) {
	type GetUserVariables struct {
		Id string `json:"id"`
	}
	type LldapUserResponseData struct {
		User LldapUser `json:"user"`
	}
	query := LldapClientQuery{
		Query:         "query GetUserDetails($id: String!) {user(userId: $id) {id email displayName firstName lastName creationDate groups {id displayName}}}",
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
	// {"query":"mutation UpdateUser($user: UpdateUserInput!) {updateUser(user: $user) {ok}}","operationName":"UpdateUser"}
	// TODO
	return diag.Errorf("Not implemented: UpdateUser")
}

func (lc *LldapClient) DeleteUser(id string) diag.Diagnostics {
	// {"query": "mutation DeleteUserQuery($user: String!) {deleteUser(userId: $user) {ok}}","operationName": "DeleteUserQuery"}
	// TODO
	return diag.Errorf("Not implemented: DeleteUser")
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
		Query:         "query ListUsersQuery($filters: RequestFilter) {users(filters: $filters) {id email displayName firstName lastName creationDate}}",
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
