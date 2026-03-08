package gcp

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var gcloudTraceState = struct {
	mu       sync.Mutex
	enabled  bool
	stepName string
	commands []string
}{
	enabled: true,
}

// SetGCloudCommandSummaryEnabled habilita ou desabilita a exibicao do resumo.
func SetGCloudCommandSummaryEnabled(enabled bool) {
	gcloudTraceState.mu.Lock()
	defer gcloudTraceState.mu.Unlock()
	gcloudTraceState.enabled = enabled
}

// BeginStepCommandTrace inicia a captura de comandos gcloud para um passo.
func BeginStepCommandTrace(stepName string) {
	gcloudTraceState.mu.Lock()
	defer gcloudTraceState.mu.Unlock()

	gcloudTraceState.stepName = stepName
	gcloudTraceState.commands = nil
}

// EndStepCommandTrace imprime o resumo dos comandos capturados e limpa o estado.
func EndStepCommandTrace() {
	gcloudTraceState.mu.Lock()
	defer gcloudTraceState.mu.Unlock()

	if !gcloudTraceState.enabled {
		gcloudTraceState.stepName = ""
		gcloudTraceState.commands = nil
		return
	}

	if len(gcloudTraceState.commands) == 0 {
		gcloudTraceState.stepName = ""
		return
	}

	if gcloudTraceState.stepName != "" {
		fmt.Printf("Comandos gcloud usados (%s):\n", gcloudTraceState.stepName)
	} else {
		fmt.Printf("Comandos gcloud usados:\n")
	}

	for i, command := range gcloudTraceState.commands {
		fmt.Printf("  %d. %s\n", i+1, command)
	}
	fmt.Printf("\n")

	gcloudTraceState.stepName = ""
	gcloudTraceState.commands = nil
}

func newGCloudCommand(args ...string) *exec.Cmd {
	appendGCloudTrace(args...)
	return exec.Command("gcloud", args...)
}

func appendGCloudTrace(args ...string) {
	gcloudTraceState.mu.Lock()
	defer gcloudTraceState.mu.Unlock()

	if !gcloudTraceState.enabled {
		return
	}

	gcloudTraceState.commands = append(gcloudTraceState.commands, formatCommandLine("gcloud", args))
}

func formatCommandLine(program string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, program)
	for _, arg := range args {
		parts = append(parts, shellQuote(arg))
	}
	return strings.Join(parts, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return `""`
	}

	if strings.ContainsAny(value, " \t\n\r\"'`$&|<>();") {
		return strconv.Quote(value)
	}

	return value
}
