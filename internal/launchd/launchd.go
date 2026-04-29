package launchd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const Label = "com.claude2-d2.daemon"

var plistTemplate = template.Must(template.New("plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{.Label}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.BinaryPath}}</string>
		<string>daemon</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>/tmp/claude2-d2.log</string>
	<key>StandardErrorPath</key>
	<string>/tmp/claude2-d2.log</string>
</dict>
</plist>
`))

func plistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", Label+".plist"), nil
}

func WritePlist(binaryPath string) error {
	path, err := plistPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := plistTemplate.Execute(&buf, struct{ Label, BinaryPath string }{Label, binaryPath}); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

func RemovePlist() error {
	path, err := plistPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func Load() error {
	path, err := plistPath()
	if err != nil {
		return err
	}
	out, err := exec.Command("launchctl", "load", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl load: %w — %s", err, bytes.TrimSpace(out))
	}
	return nil
}

func Unload() error {
	path, err := plistPath()
	if err != nil {
		return err
	}
	out, err := exec.Command("launchctl", "unload", path).CombinedOutput()
	if err != nil && !strings.Contains(string(out), "Could not find specified service") {
		return fmt.Errorf("launchctl unload: %w — %s", err, bytes.TrimSpace(out))
	}
	return nil
}
