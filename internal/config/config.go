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
	Popeye        Popeye `yaml:"popeye`
	Spinach       string
	ClearScreen   bool
	LogLevel      string
	LintLevel     string
	AllNamespaces bool

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

// ExcludedNS checks if a namespace should be excluded from the scan.
func (c *Config) ExcludedNS(n string) bool {
	return c.Popeye.Namespace.excluded(n)
}

// Init a popeye configuration from file or default if no file given.
func (c *Config) Init(f *genericclioptions.ConfigFlags) error {
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

	cfg.Popeye.LogLevel = toLogLevel(c.LogLevel)
	cfg.Popeye.LintLevel = toLintLevel(c.LintLevel)
	cfg.flags = f
	*c = cfg

	return nil
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

func toLogLevel(level string) zerolog.Level {
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

func toLintLevel(level string) int {
	switch level {
	case "ok":
		return 1
	case "info":
		return 2
	case "warn":
		return 3
	case "error":
		return 4
	default:
		return 0
	}
}

func isSet(s *string) bool {
	return s != nil && *s != ""
}
