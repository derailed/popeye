package k8s

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	restclient "k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Config tracks a kubernetes configuration.
type Config struct {
	flags        *genericclioptions.ConfigFlags
	clientConfig clientcmd.ClientConfig
	rawConfig    *clientcmdapi.Config
	restConfig   *restclient.Config
}

// NewConfig returns a new k8s config or an error if the flags are invalid.
func NewConfig(f *genericclioptions.ConfigFlags) *Config {
	return &Config{flags: f}
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

	if *c.rawConfig, err = c.clientConfig.RawConfig(); err != nil {
		return clientcmdapi.Config{}, err
	}
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
