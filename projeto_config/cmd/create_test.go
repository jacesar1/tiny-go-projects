package cmd

import (
	"io"
	"strings"
	"testing"
)

func TestCreateRejectsOptionalAPIsWithoutAll(t *testing.T) {
	createCmd := newCreateCommand()
	createCmd.SetOut(io.Discard)
	createCmd.SetErr(io.Discard)
	createCmd.SetArgs([]string{"projeto", "benner-cloud", "--optional-api", "secretmanager"})

	err := createCmd.Execute()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	if !strings.Contains(err.Error(), "so podem ser usadas com --all") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestCreateOptionalAPICommaSeparatedParse(t *testing.T) {
	createCmd := newCreateCommand()

	projectCmd, _, err := createCmd.Find([]string{"projeto"})
	if err != nil {
		t.Fatalf("could not find subcommand 'projeto': %v", err)
	}

	err = projectCmd.ParseFlags([]string{"--all", "--optional-api", "secretmanager,firestore"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	values, err := projectCmd.Flags().GetStringSlice("optional-api")
	if err != nil {
		t.Fatalf("could not read optional-api values: %v", err)
	}

	if len(values) != 2 || values[0] != "secretmanager" || values[1] != "firestore" {
		t.Fatalf("unexpected parsed values: %v", values)
	}
}
