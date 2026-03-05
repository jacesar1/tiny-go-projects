package gcp

import (
	"bytes"
	"fmt"
	"os/exec"
)

// AttachToSharedVPC vincula um projeto de serviço a um projeto host com Shared VPC
// Comando: gcloud compute shared-vpc associated-projects add <SERVICE_PROJECT> --host-project=<HOST_PROJECT>
func AttachToSharedVPC(serviceProject, hostProject string) error {
	cmd := exec.Command(
		"gcloud",
		"compute", "shared-vpc", "associated-projects", "add",
		serviceProject,
		fmt.Sprintf("--host-project=%s", hostProject),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("erro ao atachar projeto %s ao host project %s: %w\nStderr: %s", serviceProject, hostProject, err, stderr.String())
	}

	return nil
}

// GetSharedVPCStatus verifica se um projeto está atachado a uma Shared VPC
func GetSharedVPCStatus(serviceProject string) (string, error) {
	cmd := exec.Command(
		"gcloud",
		"compute", "shared-vpc", "associated-projects", "list",
		fmt.Sprintf("--filter=name=%s", serviceProject),
		"--format=value(name)",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Se retornar erro, significa que o projeto não está atachado
		return "", nil
	}

	return stdout.String(), nil
}
