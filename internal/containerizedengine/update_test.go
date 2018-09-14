package containerizedengine

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/errdefs"
	"github.com/docker/cli/cli/command"
	clitypes "github.com/docker/cli/types"
	"github.com/docker/docker/api/types"
	"gotest.tools/assert"
)

func TestActivateNoChange(t *testing.T) {
	ctx := context.Background()
	registryPrefix := "registryprefixgoeshere"
	image := &fakeImage{
		nameFunc: func() string {
			return registryPrefix + "/" + clitypes.EnterpriseEngineImage + ":engineversion"
		},
	}
	container := &fakeContainer{
		imageFunc: func(context.Context) (containerd.Image, error) {
			return image, nil
		},
		taskFunc: func(context.Context, cio.Attach) (containerd.Task, error) {
			return nil, errdefs.ErrNotFound
		},
		labelsFunc: func(context.Context) (map[string]string, error) {
			return map[string]string{}, nil
		},
	}
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return []containerd.Container{container}, nil
			},
		},
	}
	opts := clitypes.EngineInitOptions{
		EngineVersion:  "engineversiongoeshere",
		RegistryPrefix: "registryprefixgoeshere",
		ConfigFile:     "/tmp/configfilegoeshere",
		EngineImage:    clitypes.EnterpriseEngineImage,
	}

	err := client.ActivateEngine(ctx, opts, command.NewOutStream(&bytes.Buffer{}), &types.AuthConfig{}, healthfnHappy)
	assert.NilError(t, err)
}

func TestActivateDoUpdateFail(t *testing.T) {
	ctx := context.Background()
	registryPrefix := "registryprefixgoeshere"
	image := &fakeImage{
		nameFunc: func() string {
			return registryPrefix + "/ce-engine:engineversion"
		},
	}
	container := &fakeContainer{
		imageFunc: func(context.Context) (containerd.Image, error) {
			return image, nil
		},
	}
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return []containerd.Container{container}, nil
			},
			getImageFunc: func(ctx context.Context, ref string) (containerd.Image, error) {
				return nil, fmt.Errorf("something went wrong")

			},
		},
	}
	opts := clitypes.EngineInitOptions{
		EngineVersion:  "engineversiongoeshere",
		RegistryPrefix: "registryprefixgoeshere",
		ConfigFile:     "/tmp/configfilegoeshere",
		EngineImage:    clitypes.EnterpriseEngineImage,
	}

	err := client.ActivateEngine(ctx, opts, command.NewOutStream(&bytes.Buffer{}), &types.AuthConfig{}, healthfnHappy)
	assert.ErrorContains(t, err, "check for image")
	assert.ErrorContains(t, err, "something went wrong")
}

func TestDoUpdateNoVersion(t *testing.T) {
	ctx := context.Background()
	opts := clitypes.EngineInitOptions{
		EngineVersion:  "",
		RegistryPrefix: "registryprefixgoeshere",
		ConfigFile:     "/tmp/configfilegoeshere",
		EngineImage:    clitypes.EnterpriseEngineImage,
	}
	client := baseClient{}
	err := client.DoUpdate(ctx, opts, command.NewOutStream(&bytes.Buffer{}), &types.AuthConfig{}, healthfnHappy)
	assert.ErrorContains(t, err, "please pick the version you")
}

func TestDoUpdateImageMiscError(t *testing.T) {
	ctx := context.Background()
	opts := clitypes.EngineInitOptions{
		EngineVersion:  "engineversiongoeshere",
		RegistryPrefix: "registryprefixgoeshere",
		ConfigFile:     "/tmp/configfilegoeshere",
		EngineImage:    "testnamegoeshere",
	}
	client := baseClient{
		cclient: &fakeContainerdClient{
			getImageFunc: func(ctx context.Context, ref string) (containerd.Image, error) {
				return nil, fmt.Errorf("something went wrong")

			},
		},
	}
	err := client.DoUpdate(ctx, opts, command.NewOutStream(&bytes.Buffer{}), &types.AuthConfig{}, healthfnHappy)
	assert.ErrorContains(t, err, "check for image")
	assert.ErrorContains(t, err, "something went wrong")
}

func TestDoUpdatePullFail(t *testing.T) {
	ctx := context.Background()
	opts := clitypes.EngineInitOptions{
		EngineVersion:  "engineversiongoeshere",
		RegistryPrefix: "registryprefixgoeshere",
		ConfigFile:     "/tmp/configfilegoeshere",
		EngineImage:    "testnamegoeshere",
	}
	client := baseClient{
		cclient: &fakeContainerdClient{
			getImageFunc: func(ctx context.Context, ref string) (containerd.Image, error) {
				return nil, errdefs.ErrNotFound

			},
			pullFunc: func(ctx context.Context, ref string, opts ...containerd.RemoteOpt) (containerd.Image, error) {
				return nil, fmt.Errorf("pull failure")
			},
		},
	}
	err := client.DoUpdate(ctx, opts, command.NewOutStream(&bytes.Buffer{}), &types.AuthConfig{}, healthfnHappy)
	assert.ErrorContains(t, err, "unable to pull")
	assert.ErrorContains(t, err, "pull failure")
}
