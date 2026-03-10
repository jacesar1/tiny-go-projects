package cmd

import (
	"reflect"
	"strings"
	"testing"

	"projeto_config/internal/gcp"
)

func TestNormalizeOptionalAPIs(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []string
		wantErr string
	}{
		{
			name:  "accepts aliases and trims spaces",
			input: []string{" secretmanager ", "artifact-registry", "firestore"},
			want: []string{
				"secretmanager.googleapis.com",
				"artifactregistry.googleapis.com",
				"firestore.googleapis.com",
			},
		},
		{
			name:  "deduplicates values",
			input: []string{"secretmanager", "secretmanager.googleapis.com"},
			want:  []string{"secretmanager.googleapis.com"},
		},
		{
			name:    "rejects unsupported value",
			input:   []string{"bigquery"},
			wantErr: "api opcional invalida",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeOptionalAPIs(tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected normalized APIs. want=%v got=%v", tt.want, got)
			}
		})
	}
}

func TestUpdateOptionalAPICommaSeparatedParse(t *testing.T) {
	updateCmd := newUpdateCommand()

	projectCmd, _, err := updateCmd.Find([]string{"projeto"})
	if err != nil {
		t.Fatalf("could not find subcommand 'projeto': %v", err)
	}

	err = projectCmd.ParseFlags([]string{"--apis", "--optional-api", "secretmanager,firestore"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	values, err := projectCmd.Flags().GetStringSlice("optional-api")
	if err != nil {
		t.Fatalf("could not read optional-api values: %v", err)
	}

	want := []string{"secretmanager", "firestore"}
	if !reflect.DeepEqual(values, want) {
		t.Fatalf("unexpected parsed values. want=%v got=%v", want, values)
	}
}

func TestNormalizeTargetEnvironments(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []string
		wantErr string
	}{
		{
			name:  "accepts aliases and trims",
			input: []string{" dev ", "production", "qld"},
			want:  []string{"dev", "prd", "qld"},
		},
		{
			name:  "deduplicates values",
			input: []string{"dev", "dev", "prod"},
			want:  []string{"dev", "prd"},
		},
		{
			name:    "rejects unsupported value",
			input:   []string{"hml"},
			wantErr: "ambiente invalido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTargetEnvironments(tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected normalized envs. want=%v got=%v", tt.want, got)
			}
		})
	}
}

func TestUpdateEnvCommaSeparatedParse(t *testing.T) {
	updateCmd := newUpdateCommand()

	projectCmd, _, err := updateCmd.Find([]string{"projeto"})
	if err != nil {
		t.Fatalf("could not find subcommand 'projeto': %v", err)
	}

	err = projectCmd.ParseFlags([]string{"--all", "--env", "dev,qld"})
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	values, err := projectCmd.Flags().GetStringSlice("env")
	if err != nil {
		t.Fatalf("could not read env values: %v", err)
	}

	want := []string{"dev", "qld"}
	if !reflect.DeepEqual(values, want) {
		t.Fatalf("unexpected parsed env values. want=%v got=%v", want, values)
	}
}

func TestResolveOptionalAPIsForCommand(t *testing.T) {
	tests := []struct {
		name           string
		runAll         bool
		allOptional    bool
		interactive    bool
		optionalValues []string
		want           []string
		wantErr        string
	}{
		{
			name:        "all flag enables all optional APIs by default",
			runAll:      true,
			allOptional: false,
			interactive: false,
			want:        gcp.AvailableOptionalAPIs(),
		},
		{
			name:           "explicit optional list wins",
			runAll:         true,
			allOptional:    false,
			interactive:    false,
			optionalValues: []string{"secretmanager"},
			want:           []string{"secretmanager.googleapis.com"},
		},
		{
			name:        "interactive does not auto inject optional APIs",
			runAll:      true,
			allOptional: false,
			interactive: true,
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveOptionalAPIsForCommand(tt.runAll, tt.allOptional, tt.interactive, tt.optionalValues)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected optional APIs. want=%v got=%v", tt.want, got)
			}
		})
	}
}
