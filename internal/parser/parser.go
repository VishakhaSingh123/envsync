package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ── Config structures ─────────────────────────────────────────────────────────

type Config struct {
	Version      string                      `yaml:"version"`
	SourceOfTruth string                     `yaml:"source_of_truth"`
	Environments  map[string]EnvConfig       `yaml:"environments"`
	Runtimes     map[string]string           `yaml:"runtimes"`
	Secrets      SecretConfig                `yaml:"secrets"`
	Snapshots    SnapshotConfig              `yaml:"snapshots"`
}

type EnvConfig struct {
	Type    string            `yaml:"type"`    // file | aws_ssm | vault | ssh
	Path    string            `yaml:"path"`    // for file type
	Format  string            `yaml:"format"`  // dotenv | yaml | json
	Remote  RemoteConfig      `yaml:"remote"`  // for ssh/aws types
	Tags    map[string]string `yaml:"tags"`
}

type RemoteConfig struct {
	Host      string `yaml:"host"`
	User      string `yaml:"user"`
	KeyFile   string `yaml:"key_file"`
	Region    string `yaml:"region"`
	Profile   string `yaml:"profile"`
	VaultAddr string `yaml:"vault_addr"`
	VaultPath string `yaml:"vault_path"`
}

type SecretConfig struct {
	EncryptionKey string `yaml:"encryption_key_env"` // env var holding the key
	RedactedKeys  []string `yaml:"redacted_keys"`    // always mask these
}

type SnapshotConfig struct {
	Directory  string `yaml:"directory"`
	MaxKeep    int    `yaml:"max_keep"`
	Encrypted  bool   `yaml:"encrypted"`
}

// ── Loaders ───────────────────────────────────────────────────────────────────

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse %s: %w", path, err)
	}

	// Apply defaults
	if cfg.SourceOfTruth == "" {
		cfg.SourceOfTruth = ".env.example"
	}
	if cfg.Snapshots.Directory == "" {
		cfg.Snapshots.Directory = ".envsync/snapshots"
	}
	if cfg.Snapshots.MaxKeep == 0 {
		cfg.Snapshots.MaxKeep = 10
	}

	return &cfg, nil
}

func LoadEnvironment(cfg *Config, name string) (map[string]string, error) {
	envCfg, ok := cfg.Environments[name]
	if !ok {
		return nil, fmt.Errorf("environment '%s' not defined in envsync.yaml", name)
	}

	switch envCfg.Type {
	case "file", "":
		return loadFromFile(envCfg.Path, envCfg.Format)
	case "ssh":
		return loadFromSSH(envCfg)
	case "aws_ssm":
		return loadFromAWSSSM(envCfg)
	default:
		return nil, fmt.Errorf("unsupported environment type: %s", envCfg.Type)
	}
}

func LoadSourceOfTruth(cfg *Config) (map[string]string, error) {
	path := cfg.SourceOfTruth
	return loadFromFile(path, "dotenv")
}

func loadFromFile(path, format string) (map[string]string, error) {
	if format == "" {
		switch filepath.Ext(path) {
		case ".yaml", ".yml":
			format = "yaml"
		case ".json":
			format = "json"
		default:
			format = "dotenv"
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", path, err)
	}

	switch format {
	case "yaml":
		return parseYAMLEnv(data)
	case "json":
		return parseJSONEnv(data)
	default:
		return parseDotEnv(string(data))
	}
}

func parseDotEnv(content string) (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		// Strip surrounding quotes
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') ||
			(val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		result[key] = val
	}
	return result, scanner.Err()
}

func parseYAMLEnv(data []byte) (map[string]string, error) {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for k, v := range raw {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result, nil
}

func parseJSONEnv(data []byte) (map[string]string, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for k, v := range raw {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result, nil
}

// loadFromSSH is a stub — real implementation uses golang.org/x/crypto/ssh
func loadFromSSH(envCfg EnvConfig) (map[string]string, error) {
	return nil, fmt.Errorf("SSH remote loading requires a live SSH connection to %s. Run with a real server.", envCfg.Remote.Host)
}

// loadFromAWSSSM is a stub — real implementation uses aws-sdk-go-v2
func loadFromAWSSSM(envCfg EnvConfig) (map[string]string, error) {
	return nil, fmt.Errorf("AWS SSM loading requires AWS credentials and sdk. Configure your profile '%s'.", envCfg.Remote.Profile)
}

// ── Write ────────────────────────────────────────────────────────────────────

func WriteEnvironment(cfg *Config, envName string, kv map[string]string) error {
	envCfg, ok := cfg.Environments[envName]
	if !ok {
		return fmt.Errorf("environment '%s' not found", envName)
	}
	if envCfg.Type != "file" && envCfg.Type != "" {
		return fmt.Errorf("write not supported for type '%s'", envCfg.Type)
	}
	return writeDotEnv(envCfg.Path, kv)
}

func writeDotEnv(path string, kv map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprintln(w, "# Generated by envsync")
	for k, v := range kv {
		if strings.ContainsAny(v, " \t#\"'") {
			fmt.Fprintf(w, "%s=\"%s\"\n", k, v)
		} else {
			fmt.Fprintf(w, "%s=%s\n", k, v)
		}
	}
	return w.Flush()
}

// ── Scaffold ─────────────────────────────────────────────────────────────────

func ScaffoldConfig(path string) error {
	template := `# envsync.yaml — Environment Synchronization Configuration
version: "1"

# Source of truth: keys defined here are what every env MUST have
source_of_truth: .env.example

environments:
  dev:
    type: file
    path: .env.dev
    format: dotenv

  staging:
    type: file
    path: .env.staging
    format: dotenv

  production:
    type: file
    path: .env.production
    format: dotenv

  # Example: SSH remote
  # staging_remote:
  #   type: ssh
  #   remote:
  #     host: staging.example.com
  #     user: deploy
  #     key_file: ~/.ssh/id_rsa

  # Example: AWS SSM Parameter Store
  # production_aws:
  #   type: aws_ssm
  #   remote:
  #     region: us-east-1
  #     profile: myapp-prod
  #     vault_path: /myapp/production/

runtimes:
  node: "20"
  python: "3.11"
  # go: "1.21"

secrets:
  encryption_key_env: ENVSYNC_KEY   # export ENVSYNC_KEY=$(openssl rand -base64 32)
  redacted_keys:
    - PASSWORD
    - SECRET
    - TOKEN
    - KEY
    - PRIVATE

snapshots:
  directory: .envsync/snapshots
  max_keep: 10
  encrypted: true
`
	return os.WriteFile(path, []byte(template), 0644)
}
