// Package skills installs and inspects Cassor-managed runtime skill artifacts.
package skills

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rickcern44/cassor/internal/config"
	portable "github.com/rickcern44/cassor/skills/portable/cassor"
)

const manifestName = "skills-installations.json"

type InstallOptions struct {
	Agent, Scope string
	DryRun       bool
}
type Change struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}
type InstallResult struct {
	Agent       string   `json:"agent"`
	Scope       string   `json:"scope"`
	Destination string   `json:"destination"`
	Changes     []Change `json:"changes"`
}
type LifecycleResult struct {
	Agent    string   `json:"agent"`
	Scope    string   `json:"scope"`
	Changes  []Change `json:"changes"`
	Warnings []string `json:"warnings"`
}
type DoctorReport struct {
	Status   string   `json:"status"`
	Checks   []string `json:"checks"`
	Warnings []string `json:"warnings"`
}
type Manifest struct {
	SchemaVersion int            `json:"schema_version"`
	Installations []Installation `json:"installations"`
}
type Installation struct {
	Agent         string        `json:"agent"`
	Scope         string        `json:"scope"`
	Destination   string        `json:"destination"`
	CassorVersion string        `json:"cassor_version"`
	InstalledAt   string        `json:"installed_at"`
	Files         []ManagedFile `json:"files"`
}
type ManagedFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

func ManifestPath(stateDir string) string { return filepath.Join(stateDir, manifestName) }

func Install(repositoryRoot, stateDir string, options InstallOptions) (InstallResult, error) {
	if options.Agent == "all" {
		return installAll(repositoryRoot, stateDir, options)
	}
	if !supportedAgent(options.Agent) {
		return InstallResult{}, fmt.Errorf("agent %q is not supported", options.Agent)
	}
	if options.Scope == "" {
		options.Scope = "project"
	}
	if options.Scope != "project" {
		return InstallResult{}, fmt.Errorf("scope %q is not supported yet; use project", options.Scope)
	}
	destination := filepath.Join(repositoryRoot, filepath.FromSlash(destinationFor(options.Agent)))
	assets, err := assetsForAgent(options.Agent)
	if err != nil {
		return InstallResult{}, err
	}
	result := InstallResult{Agent: options.Agent, Scope: options.Scope, Destination: destination}
	manifest, err := readManifest(ManifestPath(stateDir))
	if err != nil {
		return result, err
	}
	installation := findInstallation(manifest, options.Agent, options.Scope)
	if installation != nil {
		for path, contents := range assets {
			if err := verifyManagedFile(destination, *installation, path, contents); err != nil {
				return result, err
			}
		}
		return result, nil
	}
	if _, err := os.Stat(destination); err == nil {
		return result, fmt.Errorf("skill destination %s already exists and is not Cassor-managed", destination)
	} else if !errors.Is(err, os.ErrNotExist) {
		return result, err
	}
	for path := range assets {
		result.Changes = append(result.Changes, Change{Path: filepath.Join(destination, path), Action: "create"})
	}
	sort.Slice(result.Changes, func(i, j int) bool { return result.Changes[i].Path < result.Changes[j].Path })
	if options.DryRun {
		return result, nil
	}
	if err := writeAtomically(destination, assets); err != nil {
		return result, err
	}
	managed := make([]ManagedFile, 0, len(assets))
	for path, contents := range assets {
		managed = append(managed, ManagedFile{Path: path, SHA256: hash(contents)})
	}
	sort.Slice(managed, func(i, j int) bool { return managed[i].Path < managed[j].Path })
	manifest.Installations = append(manifest.Installations, Installation{Agent: options.Agent, Scope: options.Scope, Destination: destinationFor(options.Agent), CassorVersion: "0.1.0-dev", InstalledAt: time.Now().UTC().Format(time.RFC3339Nano), Files: managed})
	if err := writeManifest(ManifestPath(stateDir), manifest); err != nil {
		return result, err
	}
	return result, nil
}

func supportedAgent(agent string) bool {
	return agent == "codex" || agent == "claude-code" || agent == "copilot"
}
func destinationFor(agent string) string {
	switch agent {
	case "codex":
		return ".agents/skills/cassor"
	case "claude-code":
		return ".claude/skills/cassor"
	case "copilot":
		return ".github/skills/cassor"
	}
	return ""
}
func installAll(root, state string, options InstallOptions) (InstallResult, error) {
	if options.Scope == "" {
		options.Scope = "project"
	}
	if options.Scope != "project" {
		return InstallResult{}, fmt.Errorf("scope %q is not supported yet; use project", options.Scope)
	}
	manifest, err := readManifest(ManifestPath(state))
	if err != nil {
		return InstallResult{}, err
	}
	result := InstallResult{Agent: "all", Scope: options.Scope}
	for _, agent := range []string{"codex", "claude-code", "copilot"} {
		assets, err := assetsForAgent(agent)
		if err != nil {
			return result, err
		}
		destination := filepath.Join(root, filepath.FromSlash(destinationFor(agent)))
		installation := findInstallation(manifest, agent, options.Scope)
		if installation != nil {
			for path, contents := range assets {
				if err := verifyManagedFile(destination, *installation, path, contents); err != nil {
					return result, err
				}
			}
			continue
		}
		if _, err := os.Stat(destination); err == nil {
			return result, fmt.Errorf("skill destination %s already exists and is not Cassor-managed", destination)
		} else if !errors.Is(err, os.ErrNotExist) {
			return result, err
		}
		for path := range assets {
			result.Changes = append(result.Changes, Change{Path: filepath.Join(destination, path), Action: "create"})
		}
	}
	if options.DryRun {
		return result, nil
	}
	for _, agent := range []string{"codex", "claude-code", "copilot"} {
		if findInstallation(manifest, agent, options.Scope) != nil {
			continue
		}
		assets, err := assetsForAgent(agent)
		if err != nil {
			return result, err
		}
		destination := filepath.Join(root, filepath.FromSlash(destinationFor(agent)))
		if err := writeAtomically(destination, assets); err != nil {
			return result, err
		}
		files := make([]ManagedFile, 0, len(assets))
		for path, contents := range assets {
			files = append(files, ManagedFile{Path: path, SHA256: hash(contents)})
		}
		manifest.Installations = append(manifest.Installations, Installation{Agent: agent, Scope: options.Scope, Destination: destinationFor(agent), CassorVersion: "0.1.0-dev", InstalledAt: time.Now().UTC().Format(time.RFC3339Nano), Files: files})
	}
	return result, writeManifest(ManifestPath(state), manifest)
}

func Status(repositoryRoot, stateDir string) (Manifest, error) {
	manifest, err := readManifest(ManifestPath(stateDir))
	if err != nil {
		return manifest, err
	}
	for _, installation := range manifest.Installations {
		destination := filepath.Join(repositoryRoot, filepath.FromSlash(installation.Destination))
		for _, file := range installation.Files {
			contents, err := os.ReadFile(filepath.Join(destination, file.Path))
			if err != nil || hash(contents) != file.SHA256 {
				return manifest, fmt.Errorf("managed skill file has changed or is missing: %s", filepath.Join(destination, file.Path))
			}
		}
	}
	return manifest, nil
}

func Update(repositoryRoot, stateDir, agent, scope string) (LifecycleResult, error) {
	manifest, err := readManifest(ManifestPath(stateDir))
	if err != nil {
		return LifecycleResult{}, err
	}
	installation := findInstallation(manifest, agent, scope)
	if installation == nil {
		return LifecycleResult{}, fmt.Errorf("no managed %s %s skill installation", scope, agent)
	}
	assets, err := assetsForAgent(agent)
	if err != nil {
		return LifecycleResult{}, err
	}
	destination := filepath.Join(repositoryRoot, filepath.FromSlash(installation.Destination))
	result := LifecycleResult{Agent: agent, Scope: scope}
	for _, file := range installation.Files {
		contents, err := os.ReadFile(filepath.Join(destination, file.Path))
		// A project skill may already match the new embedded asset when Cassor
		// itself is being developed. Accept that known target and refresh the
		// manifest, while continuing to protect unrelated local edits.
		if err != nil || (hash(contents) != file.SHA256 && hash(contents) != hash(assets[file.Path])) {
			return result, fmt.Errorf("managed skill file has changed or is missing: %s", filepath.Join(destination, file.Path))
		}
	}
	for path, contents := range assets {
		if err := writeFileAtomically(filepath.Join(destination, path), contents, 0o644); err != nil {
			return result, err
		}
		result.Changes = append(result.Changes, Change{Path: filepath.Join(destination, path), Action: "update"})
	}
	managed := make([]ManagedFile, 0, len(assets))
	for path, contents := range assets {
		managed = append(managed, ManagedFile{Path: path, SHA256: hash(contents)})
	}
	installation.Files = managed
	installation.InstalledAt = time.Now().UTC().Format(time.RFC3339Nano)
	for index := range manifest.Installations {
		if manifest.Installations[index].Agent == agent && manifest.Installations[index].Scope == scope {
			manifest.Installations[index] = *installation
		}
	}
	return result, writeManifest(ManifestPath(stateDir), manifest)
}

func Uninstall(repositoryRoot, stateDir, agent, scope string) (LifecycleResult, error) {
	manifest, err := readManifest(ManifestPath(stateDir))
	if err != nil {
		return LifecycleResult{}, err
	}
	installation := findInstallation(manifest, agent, scope)
	if installation == nil {
		return LifecycleResult{}, fmt.Errorf("no managed %s %s skill installation", scope, agent)
	}
	destination := filepath.Join(repositoryRoot, filepath.FromSlash(installation.Destination))
	result := LifecycleResult{Agent: agent, Scope: scope}
	for _, file := range installation.Files {
		path := filepath.Join(destination, file.Path)
		contents, err := os.ReadFile(path)
		if err != nil || hash(contents) != file.SHA256 {
			result.Warnings = append(result.Warnings, "preserved modified file: "+path)
			continue
		}
		if err := os.Remove(path); err != nil {
			return result, err
		}
		result.Changes = append(result.Changes, Change{Path: path, Action: "remove"})
	}
	remaining := make([]Installation, 0, len(manifest.Installations)-1)
	for _, value := range manifest.Installations {
		if value.Agent != agent || value.Scope != scope {
			remaining = append(remaining, value)
		}
	}
	manifest.Installations = remaining
	return result, writeManifest(ManifestPath(stateDir), manifest)
}

func Doctor(repositoryRoot, stateDir string) (DoctorReport, error) {
	manifest, err := Status(repositoryRoot, stateDir)
	if err != nil {
		return DoctorReport{Status: "error"}, err
	}
	report := DoctorReport{Status: "ok", Checks: []string{"Cassor project discovered", "skills manifest is valid"}}
	for _, installation := range manifest.Installations {
		report.Checks = append(report.Checks, "managed "+installation.Agent+" "+installation.Scope+" skill is current")
	}
	if len(manifest.Installations) == 0 {
		report.Warnings = append(report.Warnings, "no Cassor skills are installed")
	}
	agents, err := config.ReadAgents(config.AgentsPath(stateDir))
	if err != nil {
		report.Status = "warning"
		report.Warnings = append(report.Warnings, "Codex model routing is unavailable: "+err.Error())
	} else if err := config.ValidateCodexRouting(agents); err != nil {
		report.Status = "warning"
		report.Warnings = append(report.Warnings, "Codex model routing is incomplete: "+err.Error())
	} else {
		report.Checks = append(report.Checks, "Codex role-to-model routing is valid")
	}
	return report, nil
}

func portableAssets() (map[string][]byte, error) {
	files := map[string][]byte{}
	err := fs.WalkDir(portable.Files, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || path == "embed.go" || strings.HasPrefix(path, "references/adapters/") {
			return nil
		}
		contents, err := portable.Files.ReadFile(path)
		if err != nil {
			return err
		}
		files[path] = contents
		return nil
	})
	return files, err
}

func assetsForAgent(agent string) (map[string][]byte, error) {
	assets, err := portableAssets()
	if err != nil {
		return nil, err
	}
	if agent == "claude-code" || agent == "copilot" {
		path := "references/adapters/" + agent + ".md"
		contents, err := portable.Files.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		assets[path] = contents
	}
	return assets, nil
}
func readManifest(path string) (Manifest, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Manifest{SchemaVersion: 1}, nil
	}
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(contents, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("read skills manifest: %w", err)
	}
	if manifest.SchemaVersion != 1 {
		return Manifest{}, fmt.Errorf("unsupported skills manifest schema %d", manifest.SchemaVersion)
	}
	return manifest, nil
}
func writeManifest(path string, manifest Manifest) error {
	contents, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomically(path, append(contents, '\n'), 0o644)
}
func findInstallation(manifest Manifest, agent, scope string) *Installation {
	for index := range manifest.Installations {
		if manifest.Installations[index].Agent == agent && manifest.Installations[index].Scope == scope {
			return &manifest.Installations[index]
		}
	}
	return nil
}
func verifyManagedFile(destination string, installation Installation, path string, expected []byte) error {
	for _, file := range installation.Files {
		if file.Path == path {
			contents, err := os.ReadFile(filepath.Join(destination, path))
			if err != nil || hash(contents) != file.SHA256 {
				return fmt.Errorf("managed skill file has changed or is missing: %s", filepath.Join(destination, path))
			}
			if hash(contents) != hash(expected) {
				return fmt.Errorf("Cassor skill update available: %s; run cassor skills update", filepath.Join(destination, path))
			}
			return nil
		}
	}
	return fmt.Errorf("skills manifest is incomplete for %s", path)
}
func writeAtomically(destination string, assets map[string][]byte) error {
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	temporary, err := os.MkdirTemp(parent, ".cassor-skill-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	for path, contents := range assets {
		target := filepath.Join(temporary, path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, contents, 0o644); err != nil {
			return err
		}
	}
	return os.Rename(temporary, destination)
}
func writeFileAtomically(path string, contents []byte, permission fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, contents, permission); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
func hash(contents []byte) string { sum := sha256.Sum256(contents); return hex.EncodeToString(sum[:]) }
