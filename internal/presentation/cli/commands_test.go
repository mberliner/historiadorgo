package cli

import (
	"testing"

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