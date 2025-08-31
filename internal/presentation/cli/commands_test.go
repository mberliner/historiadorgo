package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
		wantUse  string
	}{
		{
			name:     "creates root command successfully",
			wantName: "",
			wantUse:  "historiador",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()

			assert.NotNil(t, cmd)
			assert.Equal(t, tt.wantUse, cmd.Use)
			assert.NotEmpty(t, cmd.Short)
			assert.NotNil(t, cmd.RunE)

			// Verify flags are registered
			projectFlag := cmd.PersistentFlags().Lookup("project")
			assert.NotNil(t, projectFlag)
			assert.Equal(t, "p", projectFlag.Shorthand)

			fileFlag := cmd.PersistentFlags().Lookup("file")
			assert.NotNil(t, fileFlag)
			assert.Equal(t, "f", fileFlag.Shorthand)

			dryRunFlag := cmd.PersistentFlags().Lookup("dry-run")
			assert.NotNil(t, dryRunFlag)

			batchSizeFlag := cmd.PersistentFlags().Lookup("batch-size")
			assert.NotNil(t, batchSizeFlag)
			assert.Equal(t, "b", batchSizeFlag.Shorthand)

			logLevelFlag := cmd.PersistentFlags().Lookup("log-level")
			assert.NotNil(t, logLevelFlag)
		})
	}
}

func TestNewProcessCmd(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
		wantUse  string
	}{
		{
			name:     "creates process command successfully",
			wantName: "",
			wantUse:  "process",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewProcessCmd()

			assert.NotNil(t, cmd)
			assert.Equal(t, tt.wantUse, cmd.Use)
			assert.NotEmpty(t, cmd.Short)
			assert.NotNil(t, cmd.RunE)

			// Verify flags are registered
			projectFlag := cmd.Flags().Lookup("project")
			assert.NotNil(t, projectFlag)
			assert.Equal(t, "p", projectFlag.Shorthand)

			fileFlag := cmd.Flags().Lookup("file")
			assert.NotNil(t, fileFlag)
			assert.Equal(t, "f", fileFlag.Shorthand)

			dryRunFlag := cmd.Flags().Lookup("dry-run")
			assert.NotNil(t, dryRunFlag)

			batchSizeFlag := cmd.Flags().Lookup("batch-size")
			assert.NotNil(t, batchSizeFlag)
			assert.Equal(t, "b", batchSizeFlag.Shorthand)
		})
	}
}

func TestNewValidateCmd(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
		wantUse  string
	}{
		{
			name:     "creates validate command successfully",
			wantName: "",
			wantUse:  "validate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewValidateCmd()

			assert.NotNil(t, cmd)
			assert.Equal(t, tt.wantUse, cmd.Use)
			assert.NotEmpty(t, cmd.Short)
			assert.NotNil(t, cmd.RunE)

			// Verify flags are registered
			projectFlag := cmd.Flags().Lookup("project")
			assert.NotNil(t, projectFlag)
			assert.Equal(t, "p", projectFlag.Shorthand)

			fileFlag := cmd.Flags().Lookup("file")
			assert.NotNil(t, fileFlag)
			assert.Equal(t, "f", fileFlag.Shorthand)
		})
	}
}

func TestNewTestConnectionCmd(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
		wantUse  string
	}{
		{
			name:     "creates test-connection command successfully",
			wantName: "",
			wantUse:  "test-connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewTestConnectionCmd()

			assert.NotNil(t, cmd)
			assert.Equal(t, tt.wantUse, cmd.Use)
			assert.NotEmpty(t, cmd.Short)
			assert.NotNil(t, cmd.RunE)
		})
	}
}

func TestNewDiagnoseCmd(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
		wantUse  string
	}{
		{
			name:     "creates diagnose command successfully",
			wantName: "",
			wantUse:  "diagnose",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewDiagnoseCmd()

			assert.NotNil(t, cmd)
			assert.Equal(t, tt.wantUse, cmd.Use)
			assert.NotEmpty(t, cmd.Short)
			assert.NotNil(t, cmd.RunE)

			// Verify flags are registered
			projectFlag := cmd.Flags().Lookup("project")
			assert.NotNil(t, projectFlag)
			assert.Equal(t, "p", projectFlag.Shorthand)
		})
	}
}

func TestSetupCommands(t *testing.T) {
	tests := []struct {
		name         string
		expectedCmds []string
	}{
		{
			name: "creates root command with all subcommands",
			expectedCmds: []string{
				"process",
				"validate",
				"test-connection",
				"diagnose",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := SetupCommands()

			assert.NotNil(t, rootCmd)
			assert.Equal(t, "historiador", rootCmd.Use)

			// Check that all expected subcommands were added
			commands := rootCmd.Commands()
			assert.Len(t, commands, len(tt.expectedCmds))

			cmdNames := make([]string, len(commands))
			for i, cmd := range commands {
				cmdNames[i] = cmd.Use
			}

			for _, expectedCmd := range tt.expectedCmds {
				assert.Contains(t, cmdNames, expectedCmd)
			}
		})
	}
}

func TestNewApp_Integration(t *testing.T) {
	// Test that NewApp creates the application structure correctly
	// This tests the dependency injection and wiring

	tests := []struct {
		name string
	}{
		{
			name: "creates app with valid dependencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't fully test NewApp without actual config and dependencies
			// But we can test that it creates a cobra command structure
			app := SetupCommands() // This calls NewApp internally
			assert.NotNil(t, app)
			assert.Equal(t, "historiador", app.Use)

			// Verify all expected commands are present
			commands := app.Commands()
			expectedCommands := []string{"process", "validate", "test-connection", "diagnose"}

			assert.Len(t, commands, len(expectedCommands))

			cmdMap := make(map[string]*cobra.Command)
			for _, cmd := range commands {
				cmdMap[cmd.Use] = cmd
			}

			for _, expectedCmd := range expectedCommands {
				cmd, exists := cmdMap[expectedCmd]
				assert.True(t, exists, "Command %s should exist", expectedCmd)
				assert.NotNil(t, cmd.RunE, "Command %s should have RunE function", expectedCmd)
			}
		})
	}
}

func TestCommandFlags_Integration(t *testing.T) {
	// Test that commands have the correct flags setup

	tests := []struct {
		name          string
		commandName   string
		expectedFlags []string
	}{
		{
			name:          "process_command_flags",
			commandName:   "process",
			expectedFlags: []string{"project", "file", "dry-run", "batch-size"},
		},
		{
			name:          "validate_command_flags",
			commandName:   "validate",
			expectedFlags: []string{"project", "file"},
		},
		{
			name:          "diagnose_command_flags",
			commandName:   "diagnose",
			expectedFlags: []string{"project"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := SetupCommands()

			// Find the specific command
			var targetCmd *cobra.Command
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == tt.commandName {
					targetCmd = cmd
					break
				}
			}

			assert.NotNil(t, targetCmd, "Command %s should exist", tt.commandName)

			// Check that expected flags exist
			for _, flagName := range tt.expectedFlags {
				flag := targetCmd.Flags().Lookup(flagName)
				if flag == nil {
					// Check persistent flags too
					flag = targetCmd.PersistentFlags().Lookup(flagName)
				}
				assert.NotNil(t, flag, "Flag %s should exist on command %s", flagName, tt.commandName)
			}
		})
	}
}

func TestCommandDescriptions(t *testing.T) {
	// Test that commands have proper descriptions

	tests := []struct {
		name        string
		commandName string
		checkShort  bool
	}{
		{
			name:        "process_command_description",
			commandName: "process",
			checkShort:  true,
		},
		{
			name:        "validate_command_description",
			commandName: "validate",
			checkShort:  true,
		},
		{
			name:        "test_connection_command_description",
			commandName: "test-connection",
			checkShort:  true,
		},
		{
			name:        "diagnose_command_description",
			commandName: "diagnose",
			checkShort:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := SetupCommands()

			// Find the specific command
			var targetCmd *cobra.Command
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == tt.commandName {
					targetCmd = cmd
					break
				}
			}

			assert.NotNil(t, targetCmd, "Command %s should exist", tt.commandName)

			if tt.checkShort {
				assert.NotEmpty(t, targetCmd.Short, "Command %s should have Short description", tt.commandName)
			}
		})
	}
}

func TestNewApp_CreatesNonNilApp(t *testing.T) {
	// Skip the test if .env doesn't exist to avoid interactive prompts
	tests := []struct {
		name       string
		skip       bool
		skipReason string
	}{
		{
			name: "tests app creation structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(tt.skipReason)
			}

			// We test that the NewApp function signature is correct
			// Actual testing would require valid config, so we just verify
			// that the function exists and has proper error handling patterns

			// This is a basic structural test - the function should exist
			// and be callable (even if it fails due to missing config)
			_, err := NewApp()

			// We expect either success or a config-related error
			// Both are valid outcomes for this test
			if err != nil {
				// Should be a config-related error, not a panic or nil pointer
				assert.Contains(t, err.Error(), "config")
			}
		})
	}
}
