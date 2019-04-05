package config

import (
	"io/ioutil"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	restclient "k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	defaultWidth     = 80
	defaultLogLevel  = "debug"
	defaultLintLevel = "ok"
)

// Config tracks Popeye configuration options.
type Config struct {
	Popeye      Popeye `yaml:"popeye"`
	Spinach     string
	ClearScreen bool
	LogLevel    string
	LintLevel   string
	Sections    []string

	flags        *genericclioptions.ConfigFlags
	clientConfig clientcmd.ClientConfig
	rawConfig    *clientcmdapi.Config
	restConfig   *restclient.Config
}

// New create a new Popeye configuration.
func New() *Config {
	return &Config{
		Popeye:    NewPopeye(),
		LogLevel:  defaultLogLevel,
		LintLevel: defaultLintLevel,
	}
}

// Init a popeye configuration from file or default if no file given.
func (c *Config) Init(f *genericclioptions.ConfigFlags) error {
	c.flags = f

	var cfg Config
	if len(c.Spinach) != 0 {
		f, err := ioutil.ReadFile(c.Spinach)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(f, &cfg); err != nil {
			return err
		}
	}

	cfg.Popeye.LogLevel = ToLogLevel(c.LogLevel)
	cfg.Popeye.LintLevel = ToLintLevel(c.LintLevel)
	cfg.Sections = c.Sections
	cfg.ClearScreen = c.ClearScreen
	cfg.flags = f
	*c = cfg

	return nil
}

// NodeCPULimit returns the node cpu threshold if set otherwise the default.
func (c *Config) NodeCPULimit() float64 {
	l := c.Popeye.Node.Limits.CPU
	if l == 0 {
		return defaultCPULimit
	}
	return l
}

// PodCPULimit returns the pod cpu threshold if set otherwise the default.
func (c *Config) PodCPULimit() float64 {
	l := c.Popeye.Pod.Limits.CPU
	if l == 0 {
		return defaultCPULimit
	}
	return l
}

// ExcludedNode returns excluded nodes if any.
func (c *Config) ExcludedNode(n string) bool {
	return c.Popeye.Node.excluded(n)
}

// ExcludedService returns excluded services if any.
func (c *Config) ExcludedService(s string) bool {
	return c.Popeye.Service.excluded(s)
}

// ExcludedNS checks if a namespace should be excluded from the scan.
func (c *Config) ExcludedNS(n string) bool {
	return c.Popeye.Namespace.excluded(n)
}

// RestartsLimit returns pod restarts limit.
func (c *Config) RestartsLimit() int {
	l := c.Popeye.Pod.Restarts
	if l == 0 {
		return defaultRestarts
	}
	return l
}

// PodMEMLimit returns the pod mem threshold if set otherwise the default.
func (c *Config) PodMEMLimit() float64 {
	l := c.Popeye.Pod.Limits.Memory
	if l == 0 {
		return defaultMEMLimit
	}
	return l
}

// NodeMEMLimit returns the pod mem threshold if set otherwise the default.
func (c *Config) NodeMEMLimit() float64 {
	l := c.Popeye.Node.Limits.Memory
	if l == 0 {
		return defaultMEMLimit
	}
	return l
}

// ActiveNamespace returns the desired namespace if set or all if not.
func (c *Config) ActiveNamespace() string {
	cfg, err := c.RawConfig()
	if err != nil {
		return "n/a"
	}

	if isSet(c.flags.Namespace) {
		return *c.flags.Namespace
	}

	ctx := cfg.CurrentContext
	if isSet(c.flags.Context) {
		ctx = *c.flags.Context
	}

	if c, ok := cfg.Contexts[ctx]; ok {
		return c.Namespace
	}

	return ""
}

// ActiveCluster get the current cluster name.
func (c *Config) ActiveCluster() string {
	cfg, err := c.RawConfig()
	if err != nil {
		return "n/a"
	}

	if isSet(c.flags.ClusterName) {
		return *c.flags.ClusterName
	}

	ctx := cfg.CurrentContext
	if isSet(c.flags.Context) {
		ctx = *c.flags.Context
	}

	if ctx, ok := cfg.Contexts[ctx]; ok {
		return ctx.Cluster
	}

	return "n/a"
}

// RawConfig fetch the current kubeconfig with no overrides.
func (c *Config) RawConfig() (clientcmdapi.Config, error) {
	if c.rawConfig != nil {
		return *c.rawConfig, nil
	}

	err := c.ensureClientConfig()
	if err != nil {
		return clientcmdapi.Config{}, err
	}

	raw, err := c.clientConfig.RawConfig()
	if err != nil {
		return clientcmdapi.Config{}, err
	}
	c.rawConfig = &raw

	return *c.rawConfig, nil
}

// RESTConfig fetch the current REST api-server connection.
func (c *Config) RESTConfig() (*restclient.Config, error) {
	if c.restConfig != nil {
		return c.restConfig, nil
	}

	err := c.ensureClientConfig()
	if err != nil {
		return nil, err
	}

	if c.restConfig, err = c.flags.ToRESTConfig(); err != nil {
		return nil, err
	}
	return c.restConfig, nil
}

func (c *Config) ensureClientConfig() error {
	if c.clientConfig == nil {
		c.clientConfig = c.flags.ToRawKubeConfigLoader()
	}
	return nil
}

// ToLogLevel convert a string to a level.
func ToLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Level tracks lint check level.
type Level int

const (
	// OkLevel denotes no linting issues.
	OkLevel Level = iota
	// InfoLevel denotes FIY linting issues.
	InfoLevel
	// WarnLevel denotes a warning issue.
	WarnLevel
	// ErrorLevel denotes a serious issue.
	ErrorLevel
)

// ToLintLevel convert a string to a level.
func ToLintLevel(level string) Level {
	switch level {
	case "ok":
		return OkLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return OkLevel
	}
}

func isSet(s *string) bool {
	return s != nil && *s != ""
}
