package config

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// K3sRelease
type K3sRelease struct {
	OldK8sVersion                 string `json:"old_k8s_version"`
	NewK8sVersion                 string `json:"new_k8s_version"`
	OldK8sClient                  string `json:"old_k8s_client"`
	NewK8sClient                  string `json:"new_k8s_client"`
	OldSuffix                     string `json:"old_suffix"`
	NewSuffix                     string `json:"new_suffix"`
	ReleaseBranch                 string `json:"release_branch"`
	Workspace                     string `json:"workspace"`
	NewGoVersion                  string `json:"-"`
	K3sRepoOwner                  string `json:"k3s_repo_owner"`
	SystemAgentInstallerRepoOwner string `json:"system_agent_installer_repo_owner"`
	K8sRancherURL                 string `json:"k8s_rancher_url"`
	K3sUpstreamURL                string `json:"k3s_upstream_url"`
	DryRun                        bool   `json:"dry_run"`
}

// RancherRelease
type RancherRelease struct {
	ReleaseBranch    string `json:"release_branch"`
	RancherRepoOwner string `json:"rancher_repo_owner"`
}

// RKE2
type RKE2 struct {
	Versions []string `json:"versions"`
}

// ChartsRelease
type ChartsRelease struct {
	Workspace     string   `json:"workspace"`
	ChartsRepoURL string   `json:"charts_repo_url"`
	ChartsForkURL string   `json:"charts_fork_url"`
	BranchLines   []string `json:"branch_lines"`
}

// User
type User struct {
	Email          string `json:"email"`
	GithubUsername string `json:"github_username"`
}

// K3s
type K3s struct {
	Versions map[string]K3sRelease `json:"versions"`
}

// Rancher
type Rancher struct {
	Versions map[string]RancherRelease `json:"versions"`
}

// Auth
type Auth struct {
	GithubToken        string `json:"github_token"`
	SSHKeyPath         string `json:"ssh_key_path"`
	AWSAccessKeyID     string `json:"aws_access_key_id"`
	AWSSecretAccessKey string `json:"aws_secret_access_key"`
	AWSSessionToken    string `json:"aws_session_token"`
	AWSDefaultRegion   string `json:"aws_default_region"`
}

// Config
type Config struct {
	User    *User          `json:"user"`
	K3s     *K3s           `json:"k3s"`
	Rancher *Rancher       `json:"rancher"`
	RKE2    *RKE2          `json:"rke2"`
	Charts  *ChartsRelease `json:"charts"`
	Auth    *Auth          `json:"auth"`
}

// OpenOnEditor opens the given config file on the user's default text editor.
func OpenOnEditor(configFile string) error {
	cmd := exec.Command(textEditorName(), configFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func textEditorName() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	return editor
}

// Load reads the given config file and returns a struct
// containing the necessary values to perform a release.
func Load(configFile string) (*Config, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}

	return Read(f)
}

// Read reads the given JSON file with the config and returns a struct
func Read(r io.Reader) (*Config, error) {
	var c Config
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

// ExampleConfig returns a valid JSON string with the config structure
func ExampleConfig() (string, error) {
	gopath := os.Getenv("GOPATH")

	conf := Config{
		User: &User{
			Email:          "your.name@suse.com",
			GithubUsername: "your-github-username",
		},
		K3s: &K3s{
			Versions: map[string]K3sRelease{
				"v1.x.y": {
					OldK8sVersion:                 "v1.x.z",
					NewK8sVersion:                 "v1.x.y",
					OldK8sClient:                  "v0.x.z",
					NewK8sClient:                  "v0.x.y",
					OldSuffix:                     "k3s1",
					NewSuffix:                     "k3s1",
					ReleaseBranch:                 "release-1.x",
					DryRun:                        false,
					Workspace:                     filepath.Join(gopath, "src", "github.com", "k3s-io", "kubernetes", "v1.x.z") + "/",
					SystemAgentInstallerRepoOwner: "rancher",
					K3sRepoOwner:                  "k3s-io",
					K8sRancherURL:                 "git@github.com:k3s-io/kubernetes.git",
					K3sUpstreamURL:                "git@github.com:k3s-io/k3s.git",
				},
			},
		},
		RKE2: &RKE2{
			Versions: []string{"v1.x.y"},
		},
		Rancher: &Rancher{
			Versions: map[string]RancherRelease{
				"v2.x.y": {
					ReleaseBranch:    "release/v2.x",
					RancherRepoOwner: "rancher",
				},
			},
		},
		Charts: &ChartsRelease{
			Workspace:     filepath.Join(gopath, "src", "github.com", "rancher", "charts") + "/",
			ChartsRepoURL: "https://github.com/rancher/charts",
			ChartsForkURL: "https://github.com/your-github-username/charts",
			BranchLines:   []string{"2.10", "2.9", "2.8"},
		},
		Auth: &Auth{
			GithubToken:        "YOUR_TOKEN",
			SSHKeyPath:         "path/to/your/ssh/key",
			AWSAccessKeyID:     "XXXXXXXXXXXXXXXXXXX",
			AWSSecretAccessKey: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			AWSSessionToken:    "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			AWSDefaultRegion:   "us-east-1",
		},
	}
	b, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// View prints a simplified view of the config to the standard output
func View(config *Config) error {
	tmp, err := template.New("ecm").Parse(configViewTemplate)
	if err != nil {
		return err
	}

	return tmp.Execute(os.Stdout, config)
}

const configViewTemplate = `Release config

User
	Email:           {{ .User.Email }}
	Github Username: {{ .User.GithubUsername }}

K3s {{ range $k3sVersion, $k3sValue := .K3s.Versions }}
	{{ $k3sVersion }}:
		Old K8s Version:  {{ $k3sValue.OldK8sVersion}}
		New K8s Version:  {{ $k3sValue.NewK8sVersion}}
		Old K8s Client:   {{ $k3sValue.OldK8sClient}}
		New K8s Client:   {{ $k3sValue.NewK8sClient}}
		Old Suffix:       {{ $k3sValue.OldSuffix}}
		New Suffix:       {{ $k3sValue.NewSuffix}}
		Release Branch:   {{ $k3sValue.ReleaseBranch}}
		Dry Run:          {{ $k3sValue.DryRun}}
		K3s Repo Owner:   {{ $k3sValue.K3sRepoOwner}}
		K8s Rancher URL:  {{ $k3sValue.K8sRancherURL}}
		Workspace:        {{ $k3sValue.Workspace}}
		K3s Upstream URL: {{ $k3sValue.K3sUpstreamURL}}{{ end }}

Rancher {{ range $rancherVersion, $rancherValue := .Rancher.Versions }}
	{{ $rancherVersion }}:
		Release Branch:     {{ $rancherValue.ReleaseBranch }}
		Rancher Repo Owner: {{ $rancherValue.RancherRepoOwner }}{{ end }}

RKE2{{ range .RKE2.Versions }}
	{{ . }}{{ end}}

Charts
    Workspace:     {{.Charts.Workspace}}
    ChartsRepoURL: {{.Charts.ChartsRepoURL}}
    ChartsForkURL: {{.Charts.ChartsForkURL}}
    BranchLines:     {{.Charts.BranchLines}}
`
