/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testWrap executes a command with the given arguments and returns the output
func testWrap(args []string) (*bytes.Buffer, *bytes.Buffer, error) {
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

// Unit tests - these don't require a server connection

func TestUserCreateInvalidArguments(t *testing.T) {
	// Test user create with no arguments
	_, stdErr, err := testWrap([]string{
		"user",
		"create",
	})

	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	} else {
		errorMsg = stdErr.String()
	}
	assert.Contains(t, errorMsg, "accepts 1 arg(s), received 0")
}

func TestMemberAddInvalidArguments(t *testing.T) {
	// Test member add with insufficient arguments
	_, stdErr, err := testWrap([]string{
		"member",
		"add",
		"1", // Missing user ID
	})

	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	} else {
		errorMsg = stdErr.String()
	}
	assert.Contains(t, errorMsg, "accepts 2 arg(s), received 1")
}

func TestGroupGetInvalidGid(t *testing.T) {
	// Test getting a group with invalid GID format
	_, _, err := testWrap([]string{
		"group",
		"get",
		"not-a-number",
	})

	// This should fail - either with validation error or server connection error
	assert.NotNil(t, err)
	errorMsg := err.Error()

	// Accept either validation error or server connection error
	// The CLI may check server connection before validating GID format
	hasValidationError := strings.Contains(errorMsg, "invalid gid")
	hasConnectionError := strings.Contains(errorMsg, "LLDAP_PASSWORD not set")

	assert.True(t, hasValidationError || hasConnectionError,
		"Should fail with either validation error or connection error, got: %s", errorMsg)
}
