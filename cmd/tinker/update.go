package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

const repo = "mvaliolahi/tinker"

func updateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update tinker to the latest version",
		RunE: func(_ *cobra.Command, _ []string) error {
			if _, err := exec.LookPath("git"); err != nil {
				return fmt.Errorf("git not found — required for self-update")
			}
			if _, err := exec.LookPath("go"); err != nil {
				return fmt.Errorf("go not found — required for self-update")
			}

			fmt.Println()
			fmt.Println(ui.Step(1, "Checking for updates"))

			tag, err := latestTag()
			if err != nil {
				fmt.Println(ui.Warning("Could not fetch latest version: " + err.Error()))
				return fmt.Errorf("cannot determine latest version")
			}

			fmt.Println(ui.Bullet("latest", tag))

			fmt.Println()
			fmt.Println(ui.Step(2, "Downloading "+repo+"@"+tag))

			if err := cloneAndBuild(tag); err != nil {
				fmt.Println()
				fmt.Println(ui.Warning("Clone+build failed, falling back to go install..."))
				return goInstallFallback(tag)
			}

			fmt.Println()
			fmt.Println(ui.StepDone(2, "Built and installed"))
			fmt.Println()
			fmt.Println(ui.Success("Updated to " + tag + "!"))
			return nil
		},
	}
}

func cloneAndBuild(tag string) error {
	tmp, err := os.MkdirTemp("", "tinker-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	cloneURL := fmt.Sprintf("https://github.com/%s.git", repo)
	clone := exec.Command("git", "clone", "--depth", "1", "--branch", tag, "-q", cloneURL, tmp)
	if out, err := clone.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone: %s: %w", string(out), err)
	}

	fmt.Println(ui.Dim("  Building..."))

	binDir, err := binDir()
	if err != nil {
		return err
	}

	binPath := filepath.Join(binDir, "tinker")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/tinker/")
	build.Dir = tmp
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("go build: %s: %w", string(out), err)
	}

	return nil
}

func goInstallFallback(tag string) error {
	pkg := fmt.Sprintf("github.com/%s/cmd/tinker@%s", repo, tag)
	fmt.Println(ui.Dim("  Installing " + pkg + "..."))

	cmd := exec.Command("go", "install", pkg)
	cmd.Env = append(os.Environ(), "GOFLAGS=")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.Success("Updated to " + tag + "!"))
	return nil
}

func binDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "go", "bin"), nil
}

func latestTag() (string, error) {
	if tag, err := gitLatestTag(); err == nil {
		return tag, nil
	}

	client := &http.Client{Timeout: 5 * time.Second}

	if tag, err := fetchLatestRelease(client); err == nil {
		return tag, nil
	}

	return fetchLatestTagAPI(client)
}

func gitLatestTag() (string, error) {
	url := fmt.Sprintf("https://github.com/%s.git", repo)
	cmd := exec.Command("git", "ls-remote", "--tags", "--sort=-v:refname", url)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git ls-remote: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasSuffix(line, "^{}") {
			continue
		}
		_, ref, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		return strings.TrimPrefix(ref, "refs/tags/"), nil
	}

	return "", fmt.Errorf("no tags found")
}

func fetchLatestRelease(client *http.Client) (string, error) {
	resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API status %d", resp.StatusCode)
	}

	var r struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return r.TagName, nil
}

func fetchLatestTagAPI(client *http.Client) (string, error) {
	resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/tags?per_page=1", repo))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API status %d", resp.StatusCode)
	}

	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", err
	}
	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found")
	}
	return tags[0].Name, nil
}
