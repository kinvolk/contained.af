package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// dockerProfile is an abstraction to support different configuration sets for running
// containers. More information is available here about the supported profiles and their meanings:
// https://github.com/kinvolk/container-escape-bounty/blob/master/Documentation/profiles.md
type dockerProfile string

var (
	defaultDockerProfile dockerProfile = "default-docker"
	weakDockerProfile    dockerProfile = "weak-docker"
)

var dockerProfiles = map[dockerProfile]struct{}{
	defaultDockerProfile: struct{}{},
	weakDockerProfile:    struct{}{},
}

type containerInfo struct {
	dockerImage   string
	port          string
	userns        bool
	containerid   string
	dockerProfile dockerProfile
}

func validatePort(portStr string) (nat.Port, error) {
	if portStr == "" {
		return "", nil
	}
	port, err := nat.NewPort("tcp", portStr)
	if err != nil {
		return "", err
	}
	portInt := port.Int()
	if portInt < 36100 || portInt > 36110 {
		// this is the allowed range of open ports defined in terraform config
		// https://github.com/kinvolk/container-escape-bounty/pull/19
		return "", fmt.Errorf("port not in range [36100, 36110], given: %d", portInt)
	}
	return port, nil
}

type containerOptions func(cfg *container.Config)

func withPort(port nat.Port) containerOptions {
	return func(ctrCfg *container.Config) {
		if port == "" {
			return
		}
		ctrCfg.ExposedPorts = nat.PortSet{
			port: struct{}{},
		}
	}
}

func withDockerImage(image string) containerOptions {
	return func(cfg *container.Config) {
		cfg.Image = image
		if image == "" {
			// Use default docker image when user does not provide any
			cfg.Image = defaultDockerImage
		}
	}
}

func withDockerUser(profile dockerProfile) containerOptions {
	return func(cfg *container.Config) {
		// By default, use the defaultDockerProfile.
		user := "nobody"
		if profile == weakDockerProfile {
			user = ""
		}

		cfg.User = user
	}
}

type hostOptions func(cfg *container.HostConfig) error

func withExposedPort(port nat.Port) hostOptions {
	return func(cfg *container.HostConfig) error {
		if port == "" {
			return nil
		}
		cfg.PortBindings = map[nat.Port][]nat.PortBinding{
			port: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: string(port),
				},
			},
		}
		return nil
	}
}

func withSecurityOptions(profile dockerProfile) hostOptions {
	return func(cfg *container.HostConfig) error {
		seccompConfig, ok := seccompConfigs[profile]
		if !ok {
			return fmt.Errorf("seccomp config not found for profile: %q", profile)
		}

		b := bytes.NewBuffer(nil)
		if err := json.Compact(b, []byte(seccompConfig)); err != nil {
			// this should be caught while development itself and not during runtime
			panic(fmt.Sprintf("compacting json for seccomp profile failed: %v", err))
		}
		cfg.SecurityOpt = []string{
			"no-new-privileges",
			fmt.Sprintf("seccomp=%s", b.Bytes()),
		}
		return nil
	}
}

func withHostVolumes(profile dockerProfile) hostOptions {
	return func(cfg *container.HostConfig) error {
		if profile == weakDockerProfile {
			cfg.Mounts = []mount.Mount{
				{
					Type:        mount.TypeBind,
					Source:      "/var/tmp/shared",
					Target:      "/var/tmp/shared",
					ReadOnly:    false,
					Consistency: mount.ConsistencyDefault,
				},
			}
		}
		return nil
	}
}

func withCapabilities(profile dockerProfile) hostOptions {
	return func(cfg *container.HostConfig) error {
		if profile == weakDockerProfile {
			cfg.CapAdd = []string{"NET_ADMIN", "SYS_PTRACE", "SYS_CHROOT"}
		}
		return nil
	}
}

func NewContainerConfig(opts ...containerOptions) *container.Config {
	cfg := &container.Config{
		Cmd:          []string{"sh"},
		Tty:          true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    true,
		StdinOnce:    true,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func NewContainerHostConfig(opts ...hostOptions) (*container.HostConfig, error) {
	cfg := &container.HostConfig{
		NetworkMode: "default",
		LogConfig: container.LogConfig{
			Type: "none",
		},
		Resources: container.Resources{
			PidsLimit: 5,
		},
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// startContainer starts a docker container and returns the container ID
// as well as a websocket connection to the attach endpoint.
func (h *handler) startContainer(ctrInfo *containerInfo) (*websocket.Conn, error) {
	port, err := validatePort(ctrInfo.port)
	if err != nil {
		return nil, err
	}

	ctrCfg := NewContainerConfig(
		withPort(port),
		withDockerImage(ctrInfo.dockerImage),
		withDockerUser(ctrInfo.dockerProfile),
	)

	ctrHostCfg, err := NewContainerHostConfig(
		withExposedPort(port),
		withSecurityOptions(ctrInfo.dockerProfile),
		withHostVolumes(ctrInfo.dockerProfile),
		withCapabilities(ctrInfo.dockerProfile),
	)
	if err != nil {
		return "", nil, fmt.Errorf("creating container host config: %v", err)
	}

	// using ctrCfg.Image because it is updated to be default/user given
	ctrInfo.dockerImage = ctrCfg.Image

	// pull container image if we don't already have it
	if err := h.pullImage(ctrInfo); err != nil {
		return nil, fmt.Errorf("pulling %s failed: %v", ctrCfg.Image, err)
	}

	// create the container
	r, err := h.client(ctrInfo.userns).ContainerCreate(context.Background(), ctrCfg,
		ctrHostCfg, nil, "")
	if err != nil {
		return nil, err
	}
	ctrInfo.containerid = r.ID

	// connect to the attach websocket endpoint
	header := http.Header(make(map[string][]string))
	header.Add("Origin", h.url(ctrInfo.userns).String())
	v := url.Values{
		"stdin":  []string{"1"},
		"stdout": []string{"1"},
		"stderr": []string{"1"},
		"stream": []string{"1"},
	}
	wsURL := fmt.Sprintf("ws://%s/%s/containers/%s/attach/ws?%s",
		h.url(ctrInfo.userns).Host, dockerAPIVersion, r.ID, v.Encode())
	var dialer = &websocket.Dialer{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: h.tlsConfig,
	}
	conn, _, err := dialer.Dial(wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("dialing %s with header %#v failed: %v",
			wsURL, header, err)
	}

	// start the container
	if err := h.client(ctrInfo.userns).ContainerStart(context.Background(),
		r.ID, types.ContainerStartOptions{}); err != nil {
		return conn, err
	}

	return conn, nil
}

// removeContainer removes with force a container by it's container ID.
func (h *handler) removeContainer(ctrInfo *containerInfo) error {
	if err := h.client(ctrInfo.userns).ContainerRemove(
		context.Background(),
		ctrInfo.containerid,
		types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}); err != nil {
		return err
	}

	logrus.Debugf("removed container: %s", ctrInfo.containerid)

	return nil
}

// pullImage requests a docker image if it doesn't exist already.
func (h *handler) pullImage(ctrInfo *containerInfo) error {
	exists, err := h.imageExists(ctrInfo)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	resp, err := h.client(ctrInfo.userns).ImagePull(context.Background(),
		ctrInfo.dockerImage, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	fd, isTerm := term.GetFdInfo(os.Stdout)

	return jsonmessage.DisplayJSONMessagesStream(resp, os.Stdout, fd, isTerm, nil)
}

// imageExists checks if a docker image exists.
func (h *handler) imageExists(ctrInfo *containerInfo) (bool, error) {
	_, _, err := h.client(ctrInfo.userns).ImageInspectWithRaw(
		context.Background(), ctrInfo.dockerImage)
	if err == nil {
		return true, nil
	}

	if client.IsErrNotFound(err) {
		return false, nil
	}

	return false, err
}
