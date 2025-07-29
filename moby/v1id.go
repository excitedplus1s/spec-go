package moby

import (
	"encoding/json"
	"time"

	"github.com/opencontainers/go-digest"
)

type PortSet map[Port]struct{}

type Port string

type StrSlice []string

type HealthcheckConfig struct {
	Test          []string      `json:",omitempty"`
	Interval      time.Duration `json:",omitempty"` // Interval is the time to wait between checks.
	Timeout       time.Duration `json:",omitempty"` // Timeout is the time to wait before considering the check to have hung.
	StartPeriod   time.Duration `json:",omitempty"` // The start period for the container to initialize before the retries starts to count down.
	StartInterval time.Duration `json:",omitempty"` // The interval to attempt healthchecks at during the start period
	Retries       int           `json:",omitempty"`
}

type HealthConfig = HealthcheckConfig

func (e *StrSlice) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	p := make([]string, 0, 1)
	if err := json.Unmarshal(b, &p); err != nil {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		p = append(p, s)
	}

	*e = p
	return nil
}

type Config struct {
	Hostname        string
	Domainname      string
	User            string
	AttachStdin     bool
	AttachStdout    bool
	AttachStderr    bool
	ExposedPorts    PortSet `json:",omitempty"`
	Tty             bool
	OpenStdin       bool
	StdinOnce       bool
	Env             []string
	Cmd             StrSlice
	Healthcheck     *HealthConfig `json:",omitempty"`
	ArgsEscaped     bool          `json:",omitempty"`
	Image           string
	Volumes         map[string]struct{}
	WorkingDir      string
	Entrypoint      StrSlice
	NetworkDisabled bool   `json:",omitempty"`
	MacAddress      string `json:",omitempty"`
	OnBuild         []string
	Labels          map[string]string
	StopSignal      string   `json:",omitempty"`
	StopTimeout     *int     `json:",omitempty"`
	Shell           StrSlice `json:",omitempty"`
}

type V1Image struct {
	ID              string     `json:"id,omitempty"`
	Parent          string     `json:"parent,omitempty"`
	Comment         string     `json:"comment,omitempty"`
	Created         *time.Time `json:"created"`
	Container       string     `json:"container,omitempty"`
	ContainerConfig Config     `json:"container_config,omitempty"`
	DockerVersion   string     `json:"docker_version,omitempty"`
	Author          string     `json:"author,omitempty"`
	Config          *Config    `json:"config,omitempty"`
	Architecture    string     `json:"architecture,omitempty"`
	Variant         string     `json:"variant,omitempty"`
	OS              string     `json:"os,omitempty"`
	Size            int64      `json:",omitempty"`
}

func rawJSON(value interface{}) *json.RawMessage {
	jsonval, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return (*json.RawMessage)(&jsonval)
}

func CreateID(v1Image V1Image, layerID digest.Digest, parent digest.Digest) (digest.Digest, error) {
	v1Image.ID = ""
	v1JSON, err := json.Marshal(v1Image)
	if err != nil {
		return "", err
	}

	var config map[string]*json.RawMessage
	if err := json.Unmarshal(v1JSON, &config); err != nil {
		return "", err
	}

	// FIXME: note that this is slightly incompatible with RootFS logic
	config["layer_id"] = rawJSON(layerID)
	if parent != "" {
		config["parent"] = rawJSON(parent)
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	return digest.FromBytes(configJSON), nil
}
