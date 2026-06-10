package bootstrap

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

func TestAWSEmulatorContainerTerminateIsIdempotent(t *testing.T) {
	container := &fakeTestcontainersContainer{}
	emulator := &AWSEmulatorContainer{Container: container}

	require.NoError(t, emulator.Terminate(context.Background()))
	require.NoError(t, emulator.Terminate(context.Background()))

	assert.Equal(t, 1, container.terminateCalls)
	assert.Nil(t, emulator.Container)
}

func TestAWSEmulatorContainerTerminateRetainsContainerAfterFailure(t *testing.T) {
	termErr := errors.New("boom")
	container := &fakeTestcontainersContainer{terminateErrs: []error{termErr, nil}}
	emulator := &AWSEmulatorContainer{Container: container}

	require.ErrorIs(t, emulator.Terminate(context.Background()), termErr)
	assert.NotNil(t, emulator.Container)

	require.NoError(t, emulator.Terminate(context.Background()))
	assert.Equal(t, 2, container.terminateCalls)
	assert.Nil(t, emulator.Container)
}

type fakeTestcontainersContainer struct {
	terminateCalls int
	terminateErrs  []error
}

func (c *fakeTestcontainersContainer) GetContainerID() string {
	return ""
}

func (c *fakeTestcontainersContainer) Endpoint(context.Context, string) (string, error) {
	return "", nil
}

func (c *fakeTestcontainersContainer) PortEndpoint(context.Context, nat.Port, string) (string, error) {
	return "", nil
}

func (c *fakeTestcontainersContainer) Host(context.Context) (string, error) {
	return "", nil
}

func (c *fakeTestcontainersContainer) Inspect(context.Context) (*dockercontainer.InspectResponse, error) {
	return nil, nil
}

func (c *fakeTestcontainersContainer) MappedPort(context.Context, nat.Port) (nat.Port, error) {
	return "", nil
}

func (c *fakeTestcontainersContainer) Ports(context.Context) (nat.PortMap, error) {
	return nil, nil
}

func (c *fakeTestcontainersContainer) SessionID() string {
	return ""
}

func (c *fakeTestcontainersContainer) IsRunning() bool {
	return false
}

func (c *fakeTestcontainersContainer) Start(context.Context) error {
	return nil
}

func (c *fakeTestcontainersContainer) Stop(context.Context, *time.Duration) error {
	return nil
}

func (c *fakeTestcontainersContainer) Terminate(context.Context, ...testcontainers.TerminateOption) error {
	c.terminateCalls++
	if len(c.terminateErrs) == 0 {
		return nil
	}
	err := c.terminateErrs[0]
	c.terminateErrs = c.terminateErrs[1:]
	return err
}

func (c *fakeTestcontainersContainer) Logs(context.Context) (io.ReadCloser, error) {
	return nil, nil
}

func (c *fakeTestcontainersContainer) FollowOutput(testcontainers.LogConsumer) {}

func (c *fakeTestcontainersContainer) StartLogProducer(context.Context, ...testcontainers.LogProductionOption) error {
	return nil
}

func (c *fakeTestcontainersContainer) StopLogProducer() error {
	return nil
}

func (c *fakeTestcontainersContainer) Name(context.Context) (string, error) {
	return "", nil
}

func (c *fakeTestcontainersContainer) State(context.Context) (*dockercontainer.State, error) {
	return nil, nil
}

func (c *fakeTestcontainersContainer) Networks(context.Context) ([]string, error) {
	return nil, nil
}

func (c *fakeTestcontainersContainer) NetworkAliases(context.Context) (map[string][]string, error) {
	return nil, nil
}

func (c *fakeTestcontainersContainer) Exec(context.Context, []string, ...tcexec.ProcessOption) (int, io.Reader, error) {
	return 0, nil, nil
}

func (c *fakeTestcontainersContainer) ContainerIP(context.Context) (string, error) {
	return "", nil
}

func (c *fakeTestcontainersContainer) ContainerIPs(context.Context) ([]string, error) {
	return nil, nil
}

func (c *fakeTestcontainersContainer) CopyToContainer(context.Context, []byte, string, int64) error {
	return nil
}

func (c *fakeTestcontainersContainer) CopyDirToContainer(context.Context, string, string, int64) error {
	return nil
}

func (c *fakeTestcontainersContainer) CopyFileToContainer(context.Context, string, string, int64) error {
	return nil
}

func (c *fakeTestcontainersContainer) CopyFileFromContainer(context.Context, string) (io.ReadCloser, error) {
	return nil, nil
}

func (c *fakeTestcontainersContainer) GetLogProductionErrorChannel() <-chan error {
	return nil
}
