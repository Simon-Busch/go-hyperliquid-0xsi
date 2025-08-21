package hyperliquid

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// createTempPythonScript creates a temporary Python script file and returns its path
func createTempPythonScript(scriptContent, scriptName string) (string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "hyperliquid_python_bridge")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create the script file
	scriptPath := filepath.Join(tempDir, scriptName)
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		os.RemoveAll(tempDir) // Clean up on error
		return "", fmt.Errorf("failed to write script file: %w", err)
	}

	return scriptPath, nil
}

// findPythonExecutable searches for a Python executable in common locations
func findPythonExecutable() string {
	// Try common Python executable names in order of preference
	pythonCmds := []string{"python3", "python", "python3.11", "python3.10", "python3.9", "python3.8"}

	for _, cmd := range pythonCmds {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}

	// Try common installation paths
	commonPaths := []string{
		"/usr/bin/python3",
		"/usr/local/bin/python3",
		"/opt/homebrew/bin/python3",
		"/usr/bin/python",
		"/usr/local/bin/python",
	}

	for _, path := range commonPaths {
		if _, err := exec.Command(path, "--version").Output(); err == nil {
			return path
		}
	}

	return ""
}

// callPythonBridge is a helper function to call Python scripts with embedded content
func callPythonBridge(scriptContent string, scriptName string, args ...string) ([]byte, error) {
	// Find Python executable
	pythonCmd := findPythonExecutable()
	if pythonCmd == "" {
		return nil, fmt.Errorf("python executable not found in PATH")
	}

	// Create temporary script
	scriptPath, err := createTempPythonScript(scriptContent, scriptName)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary script: %w", err)
	}

	// Clean up the temporary file after execution
	defer func() {
		os.RemoveAll(filepath.Dir(scriptPath))
	}()

	// Build command with script path and arguments
	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.Command(pythonCmd, cmdArgs...)

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to call Python bridge: %w\nOutput: %s", err, string(output))
	}

	return output, nil
}
