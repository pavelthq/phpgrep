package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEnd2End(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	phpgrepBin := filepath.Join(wd, "phpgrep.exe")

	out, err := exec.Command("go", "build", "-race", "-o", phpgrepBin, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("build phpgrep: %v: %s", err, out)
	}

	type patternTest struct {
		pattern string
		filters []string
		matches []string
	}
	tests := []struct {
		name  string
		tests []patternTest
	}{
		{
			name: "filter",
			tests: []patternTest{
				// Test '=' for strings.
				{
					pattern: `define($name, $_)`,
					filters: []string{`name="FOO"`},
					matches: []string{`file.php:3: define("FOO", 1)`},
				},
				{
					pattern: `define($name, $_)`,
					filters: []string{`name='FOO'`},
					matches: []string{`file.php:4: define('FOO', 2)`},
				},
				{
					pattern: `define($name, $_)`,
					filters: []string{`name='FOO','BAR'`},
					matches: []string{
						`file.php:4: define('FOO', 2)`,
						`file.php:5: define('BAR', 3)`,
					},
				},

				// Test `~` for strings.
				{
					pattern: `define($name, $_)`,
					filters: []string{`name~"FOO"`},
					matches: []string{`file.php:3: define("FOO", 1)`},
				},
				{
					pattern: `define($name, $_)`,
					filters: []string{`name~^"FOO"$`},
					matches: []string{`file.php:3: define("FOO", 1)`},
				},
				{
					pattern: `define($name, $_)`,
					filters: []string{`name~^."FOO"$`},
				},
				{
					pattern: `define($name, $_)`,
					filters: []string{`name~^.."FOO".$`},
				},

				// Test '=' for ints.
				{
					pattern: `define($_, $v)`,
					filters: []string{`v=2`},
					matches: []string{
						`file.php:4: define('FOO', 2)`,
					},
				},
				{
					pattern: `define($_, $v)`,
					filters: []string{`v=1,2`},
					matches: []string{
						`file.php:3: define("FOO", 1)`,
						`file.php:4: define('FOO', 2)`,
					},
				},

				// Test '!=' for ints.
				{
					pattern: `define($_, $v)`,
					filters: []string{`v!=2`},
					matches: []string{
						`file.php:3: define("FOO", 1)`,
						`file.php:5: define('BAR', 3)`,
					},
				},
				{
					pattern: `define($_, $v)`,
					filters: []string{`v!=3,2`},
					matches: []string{
						`file.php:3: define("FOO", 1)`,
					},
				},
			},
		},
	}

	for _, test := range tests {
		testName := test.name
		patternTests := test.tests
		t.Run(test.name, func(t *testing.T) {
			target := filepath.Join("testdata", testName)
			if err := os.Chdir(target); err != nil {
				t.Fatalf("chdir to test: %v", err)
			}

			for _, test := range patternTests {
				phpgrepArgs := []string{".", test.pattern}
				phpgrepArgs = append(phpgrepArgs, test.filters...)
				out, err := exec.Command(phpgrepBin, phpgrepArgs...).CombinedOutput()
				if err != nil {
					if getExitCode(err) == 1 && len(test.matches) == 0 {
						// OK: exit code 1 means "no matches".
						return
					}
					t.Fatalf("run phpgrep: %v: %s", err, out)
				}
				have := strings.Split(strings.TrimSpace(string(out)), "\n")
				want := test.matches
				if diff := cmp.Diff(want, have); diff != "" {
					t.Errorf("output mismatch (+have -want):\n%s", diff)
				}
			}

			if err := os.Chdir(wd); err != nil {
				t.Fatalf("chdir back: %v", err)
			}
		})
	}
}

func getExitCode(err error) int {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return -1
}
