package extensions

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: adm extensions <command>")
	}

	switch args[0] {
	case "list":
		return runList()
	case "upgrade":
		return runUpgrade()
	case "help", "-h", "--help":
		Usage()
		return nil
	default:
		return fmt.Errorf("unknown extensions command: %s", args[0])
	}
}

func Usage() {
	fmt.Println("usage:")
	fmt.Println("  adm extensions <command>")
	fmt.Println()
	fmt.Println("available commands:")
	fmt.Println("  list     show MediaWiki extension names and versions")
	fmt.Println("  upgrade  refresh w/extensions from mwz/extensions/extensions.yaml")
}

type Extension struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Repo    string `json:"repo"`
	Tag     string `json:"tag"`
}

func runList() error {
	rootDir, err := repoRootFrom(".")
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(filepath.Join(rootDir, "w", "extensions"))
	if err != nil {
		return err
	}

	rows := make([]Extension, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extensionJSON := filepath.Join(rootDir, "w", "extensions", entry.Name(), "extension.json")
		data, err := os.ReadFile(extensionJSON)
		if err != nil {
			continue
		}

		var info Extension
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}
		if info.Name == "" || info.Version == "" {
			continue
		}

		rows = append(rows, info)
	}

	mwzEntries, err := parseYAML(filepath.Join(rootDir, "mwz", "extensions", "extensions.yaml"))
	if err != nil {
		return err
	}
	rows = mergeExtensions(rows, mwzEntries)

	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].Name) < strings.ToLower(rows[j].Name)
	})

	return printExtensionRows(rows)
}

func runUpgrade() error {
	rootDir, err := repoRootFrom(".")
	if err != nil {
		return err
	}

	entries, err := parseYAML(filepath.Join(rootDir, "mwz", "extensions", "extensions.yaml"))
	if err != nil {
		return err
	}

	extensionsDir := filepath.Join(rootDir, "w", "extensions")
	for _, entry := range entries {
		targetDir := filepath.Join(extensionsDir, entry.Name)
		extensionJSON := filepath.Join(targetDir, "extension.json")
		desiredVersion := trimTagPrefix(entry.Tag)

		targetStat, statErr := os.Lstat(targetDir)
		if statErr != nil || !targetStat.IsDir() {
			if statErr == nil {
				if err := os.RemoveAll(targetDir); err != nil {
					return err
				}
			}
			fmt.Printf("[%s] missing directory, cloning %s...\n", entry.Name, entry.Tag)
			if err := cloneExtensionRepo(entry.Repo, entry.Tag, targetDir); err != nil {
				return err
			}
			continue
		}

		currentVersion := ""
		if data, err := os.ReadFile(extensionJSON); err == nil {
			var info Extension
			if err := json.Unmarshal(data, &info); err == nil {
				currentVersion = info.Version
			}
		}

		if trimTagPrefix(currentVersion) == desiredVersion {
			fmt.Printf("[%s] up to date (%s)\n", entry.Name, desiredVersion)
			continue
		}

		fmt.Printf(
			"[%s] version mismatch (current: %s, desired: %s); refreshing...\n",
			entry.Name,
			blankToDash(currentVersion),
			desiredVersion,
		)
		if err := os.RemoveAll(targetDir); err != nil {
			return err
		}
		if err := cloneExtensionRepo(entry.Repo, entry.Tag, targetDir); err != nil {
			return err
		}
	}

	return nil
}

func mergeExtensions(base []Extension, extra []Extension) []Extension {
	byName := make(map[string]Extension, len(base)+len(extra))

	for _, row := range base {
		byName[strings.ToLower(row.Name)] = row
	}
	for _, row := range extra {
		key := strings.ToLower(row.Name)
		current, ok := byName[key]
		if !ok {
			byName[key] = row
			continue
		}
		if current.Name == "" {
			current.Name = row.Name
		}
		if current.Version == "" {
			current.Version = row.Version
		}
		if current.Repo == "" {
			current.Repo = row.Repo
		}
		if current.Tag == "" {
			current.Tag = row.Tag
		}
		byName[key] = current
	}

	rows := make([]Extension, 0, len(byName))
	for _, row := range byName {
		rows = append(rows, row)
	}
	return rows
}

func printExtensionRows(rows []Extension) error {
	type renderedRow struct {
		name     string
		version  string
		tag      string
		mismatch bool
	}

	rendered := make([]renderedRow, 0, len(rows))
	nameWidth := len("NAME")
	versionWidth := len("VERSION")
	tagWidth := len("TAG")

	for _, row := range rows {
		version := row.Version
		if version == "" {
			version = "-"
		}
		tag := row.Tag
		if tag == "" {
			tag = "-"
		}
		mismatch := row.Version != "" && row.Tag != "" && row.Version != row.Tag && row.Version != trimTagPrefix(row.Tag)
		rendered = append(rendered, renderedRow{
			name:     row.Name,
			version:  version,
			tag:      tag,
			mismatch: mismatch,
		})
		nameWidth = maxInt(nameWidth, len(row.Name))
		versionWidth = maxInt(versionWidth, len(version))
		tagWidth = maxInt(tagWidth, len(tag))
	}

	fmt.Printf("%-*s  %-*s  %-*s\n", nameWidth, "NAME", versionWidth, "VERSION", tagWidth, "TAG")
	for _, row := range rendered {
		line := fmt.Sprintf("%-*s  %-*s  %-*s", nameWidth, row.name, versionWidth, row.version, tagWidth, row.tag)
		if row.mismatch {
			fmt.Printf("\x1b[31m%s\x1b[0m\n", line)
			continue
		}
		fmt.Println(line)
	}
	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func cloneExtensionRepo(repo string, tag string, targetDir string) error {
	if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
		return err
	}

	cmd := exec.Command("git", "clone", "--depth=1", "-b", tag, repo, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func blankToDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func parseYAML(filePath string) ([]Extension, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var rows []Extension
	var current *Extension
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "- name:") {
			if current != nil && current.Name != "" && current.Tag != "" && current.Repo != "" {
				rows = append(rows, *current)
			}
			current = &Extension{
				Name: trimQuotes(strings.TrimSpace(strings.TrimPrefix(line, "- name:"))),
			}
			continue
		}

		if current == nil {
			continue
		}

		switch {
		case strings.HasPrefix(line, "repo:"):
			current.Repo = trimQuotes(strings.TrimSpace(strings.TrimPrefix(line, "repo:")))
		case strings.HasPrefix(line, "tag:"):
			current.Tag = trimQuotes(strings.TrimSpace(strings.TrimPrefix(line, "tag:")))
		}
	}

	if current != nil && current.Name != "" && current.Tag != "" && current.Repo != "" {
		rows = append(rows, *current)
	}

	return rows, nil
}

func trimTagPrefix(tag string) string {
	return strings.TrimPrefix(tag, "v")
}

func repoRootFrom(cwd string) (string, error) {
	dir := cwd
	for {
		wExt := filepath.Join(dir, "w", "extensions")
		mwzYaml := filepath.Join(dir, "mwz", "extensions", "extensions.yaml")
		if st, err := os.Stat(wExt); err == nil && st.IsDir() {
			if _, err := os.Stat(mwzYaml); err == nil {
				return dir, nil
			}
		}
		if st, err := os.Stat(mwzYaml); err == nil && !st.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find repository root from current directory")
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
