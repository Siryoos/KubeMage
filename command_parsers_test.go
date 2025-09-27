package main

import "testing"

func TestParseGenDeployCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    DeploymentOptions
		wantErr bool
	}{
		{
			name:  "minimal",
			input: "/gen-deploy api --image ghcr.io/example/api:latest",
			want: DeploymentOptions{
				Name:  "api",
				Image: "ghcr.io/example/api:latest",
			},
		},
		{
			name:  "with replicas and port",
			input: "/gen-deploy web --image nginx:1.27 --replicas 3 --port 8080",
			want: DeploymentOptions{
				Name:     "web",
				Image:    "nginx:1.27",
				Replicas: 3,
				Port:     8080,
			},
		},
		{
			name:    "missing image",
			input:   "/gen-deploy svc",
			wantErr: true,
		},
		{
			name:    "unknown flag",
			input:   "/gen-deploy svc --foo bar",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGenDeployCommand(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none (options: %+v)", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Name != tt.want.Name {
				t.Fatalf("expected name %q, got %q", tt.want.Name, got.Name)
			}
			if got.Image != tt.want.Image {
				t.Fatalf("expected image %q, got %q", tt.want.Image, got.Image)
			}
			if got.Replicas != tt.want.Replicas {
				t.Fatalf("expected replicas %d, got %d", tt.want.Replicas, got.Replicas)
			}
			if got.Port != tt.want.Port {
				t.Fatalf("expected port %d, got %d", tt.want.Port, got.Port)
			}
		})
	}
}

func TestParseGenHelmCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    HelmChartOptions
		wantErr bool
	}{
		{
			name:  "defaults",
			input: "/gen-helm telemetry",
			want:  HelmChartOptions{Name: "telemetry"},
		},
		{
			name:  "with options",
			input: "/gen-helm billing --description payments --version 1.2.3 --app-version 9.9.9",
			want: HelmChartOptions{
				Name:        "billing",
				Description: "payments",
				Version:     "1.2.3",
				AppVersion:  "9.9.9",
			},
		},
		{
			name:    "unknown flag",
			input:   "/gen-helm svc --foo bar",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGenHelmCommand(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none (options: %+v)", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Name != tt.want.Name {
				t.Fatalf("expected name %q, got %q", tt.want.Name, got.Name)
			}
			if got.Description != tt.want.Description {
				t.Fatalf("expected description %q, got %q", tt.want.Description, got.Description)
			}
			if got.Version != tt.want.Version {
				t.Fatalf("expected version %q, got %q", tt.want.Version, got.Version)
			}
			if got.AppVersion != tt.want.AppVersion {
				t.Fatalf("expected app version %q, got %q", tt.want.AppVersion, got.AppVersion)
			}
		})
	}
}
