package install

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestInstallShellEnvironmentPrecedence(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX installer is not exercised on Windows")
	}

	probe := strings.Replace(
		installShell,
		`do_install "$@" || exit 1`,
		`printf '%s|%s|%s|%s\n' "$INSTALL_MODE" "$INSTALL_APPS" "$ARTIFACT_BASE_URL" "$DEBUG"`,
		1,
	)
	if probe == installShell {
		t.Fatal("installer entrypoint not found")
	}

	tests := []struct {
		name string
		env  []string
		args []string
		want string
	}{
		{
			name: "documented variables",
			env: []string{
				"SYFTBOX_INSTALL_MODE=download-only",
				"SYFTBOX_INSTALL_APPS=alpha,beta",
				"SYFTBOX_ARTIFACT_BASE_URL=https://artifacts.example",
				"SYFTBOX_DEBUG=1",
			},
			want: "download-only|alpha,beta|https://artifacts.example|1",
		},
		{
			name: "documented variables beat legacy variables",
			env: []string{
				"SYFTBOX_INSTALL_MODE=setup-only",
				"INSTALL_MODE=interactive",
				"SYFTBOX_INSTALL_APPS=namespaced",
				"INSTALL_APPS=legacy",
				"SYFTBOX_ARTIFACT_BASE_URL=https://namespaced.example",
				"ARTIFACT_BASE_URL=https://legacy.example",
				"SYFTBOX_DEBUG=1",
				"DEBUG=0",
			},
			want: "setup-only|namespaced|https://namespaced.example|1",
		},
		{
			name: "command line beats environment",
			env: []string{
				"SYFTBOX_INSTALL_MODE=setup-only",
				"SYFTBOX_INSTALL_APPS=environment",
			},
			args: []string{"--download-only", "--apps=command-line", "--debug"},
			want: "download-only|command-line|https://syftbox.net|1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("sh", append([]string{"-s", "--"}, tt.args...)...)
			cmd.Stdin = strings.NewReader(probe)
			cmd.Env = append(os.Environ(), tt.env...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("installer probe failed: %v\n%s", err, out)
			}
			if got := strings.TrimSpace(string(out)); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
