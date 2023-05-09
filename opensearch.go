// inspired by https://github.com/opensearch-project/opensearch-testcontainers/blob/release-2.0.0/src/main/java/org/opensearch/testcontainers/OpensearchContainer.java

package opensearch

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"time"
)

const (
	defaultUser     = "admin"
	defaultPassword = "admin"
	defaultHttpPort = "9200/tcp"
	defaultTcpPort  = "9300/tcp"
)

// Container represents the opensearch container type used in the module
type Container struct {
	testcontainers.Container
	disableSecurity bool
}

// RunContainer creates an instance of the opensearch container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image: "opensearchproject/opensearch:2.5.0",
		Env: map[string]string{
			"DISABLE_SECURITY_PLUGIN": "true",
			"discovery.type":          "single-node",
		},
		ExposedPorts: []string{defaultHttpPort, defaultTcpPort},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	disableSecurity := isSecurityDisabled(&genericContainerReq)
	if !disableSecurity {
		delete(genericContainerReq.Env, "DISABLE_SECURITY_PLUGIN")
	}

	genericContainerReq.WaitingFor = createWaitStrategyFor(disableSecurity)

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &Container{
		Container:       container,
		disableSecurity: disableSecurity,
	}, nil
}

func createWaitStrategyFor(disableSecurity bool) wait.Strategy {
	if !disableSecurity {
		return wait.NewHTTPStrategy("/").
			WithTLS(true).
			WithAllowInsecure(true).
			WithPort(defaultHttpPort).
			WithBasicAuth(defaultUser, defaultPassword).
			WithStatusCodeMatcher(func(status int) bool { return status == http.StatusOK || status == http.StatusUnauthorized }).
			WithStartupTimeout(10 * time.Second).
			WithStartupTimeout(5 * time.Minute)
	}
	return wait.NewHTTPStrategy("/").
		WithPort(defaultHttpPort).
		WithStatusCodeMatcher(func(status int) bool { return status == http.StatusOK }).
		WithStartupTimeout(10 * time.Second).
		WithStartupTimeout(5 * time.Minute)
}

func isSecurityDisabled(req *testcontainers.GenericContainerRequest) bool {
	return req.Env["DISABLE_SECURITY_PLUGIN"] == "true"
}

func WithSecurityEnabled() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["DISABLE_SECURITY_PLUGIN"] = "false"
	}
}

func (c *Container) GetHttpHostAddress(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := c.MappedPort(ctx, defaultHttpPort)
	if err != nil {
		return "", err
	}

	if c.disableSecurity {
		return "http://" + host + ":" + port.Port(), nil
	}
	return "https://" + host + ":" + port.Port(), nil
}

func (c *Container) IsSecurityEnabled() bool {
	return !c.disableSecurity
}

func (c *Container) GetUserName() string {
	return defaultUser
}

func (c *Container) GetPassword() string {
	return defaultPassword
}
