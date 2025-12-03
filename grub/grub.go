package grub

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const GrubConfigPath = "/etc/default/grub"
const GrubCfgPath = "/boot/grub/grub.cfg"
const CustomProxyScriptPath = "/etc/grub.d/42_custom_proxy"

// MenuEntry represents a grub menu entry.
type MenuEntry struct {
	Title   string
	Content string
}

// ParseGrubCfg parses the grub.cfg file and returns a list of menu entries.
func ParseGrubCfg() ([]MenuEntry, error) {
	file, err := os.Open(GrubCfgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var menuEntries []MenuEntry
	var currentEntry *MenuEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "menuentry") {
			if currentEntry != nil {
				menuEntries = append(menuEntries, *currentEntry)
			}
			parts := strings.SplitN(line, "'", 3)
			if len(parts) < 2 {
				// Invalid menuentry line, skip it
				continue
			}
			currentEntry = &MenuEntry{
				Title:   parts[1],
				Content: line + "\n",
			}
		} else if currentEntry != nil {
			currentEntry.Content += line + "\n"
			if trimmedLine == "}" {
				menuEntries = append(menuEntries, *currentEntry)
				currentEntry = nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return menuEntries, nil
}

// WriteGrubConfig writes the given content to the grub config file.
// It uses pkexec to get root privileges to write to the file.
func WriteGrubConfig(content string) error {
	// Create a temporary file to write the content to
	tmpfile, err := os.CreateTemp("", "grub-config-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}
	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Use pkexec to copy the temporary file to the grub config path
	cmd := exec.Command("/usr/bin/pkexec", "cp", tmpfile.Name(), GrubConfigPath)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run pkexec: %s", out.String())
	}

	return nil
}

// RunGrubMkconfig runs the grub-mkconfig command with pkexec.
func RunGrubMkconfig() (string, error) {
	cmd := exec.Command("/usr/bin/pkexec", "grub-mkconfig", "-o", GrubCfgPath)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("failed to run grub-mkconfig: %w", err)
	}
	return out.String(), nil
}

// WriteCustomProxyScript writes the menu entries to a custom grub script.
func WriteCustomProxyScript(menuEntries []MenuEntry) error {
	var builder strings.Builder
	builder.WriteString("#!/bin/sh\n")
	builder.WriteString("exec tail -n +3 $0\n")
	for _, entry := range menuEntries {
		builder.WriteString(entry.Content)
	}

	// Create a temporary file to write the content to
	tmpfile, err := os.CreateTemp("", "grub-proxy-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(builder.String()); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}
	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Use pkexec to copy the temporary file to the grub config path
	cmd := exec.Command("/usr/bin/pkexec", "cp", tmpfile.Name(), CustomProxyScriptPath)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run pkexec: %s", out.String())
	}

	// Make the script executable
	cmd = exec.Command("/usr/bin/pkexec", "chmod", "+x", CustomProxyScriptPath)
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to make script executable: %s", out.String())
	}

	return nil
}

// DisableOsProber disables the os prober in the grub config.
func DisableOsProber() error {
	// Read the grub config file
	grubBytes, err := os.ReadFile(GrubConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read grub config: %w", err)
	}

	// Add or update the GRUB_DISABLE_OS_PROBER setting
	content := string(grubBytes)
	if strings.Contains(content, "GRUB_DISABLE_OS_PROBER") {
		content = strings.Replace(content, "GRUB_DISABLE_OS_PROBER=false", "GRUB_DISABLE_OS_PROBER=true", 1)
	} else {
		content += "\nGRUB_DISABLE_OS_PROBER=true"
	}

	// Write the changes back to the file
	return WriteGrubConfig(content)
}
