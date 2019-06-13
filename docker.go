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
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type containerInfo struct {
	dockerImage string
	port        string
}

func generatePort(portStr string) (nat.PortSet, error) {
	if portStr == "" {
		return nil, nil
	}
	port, err := nat.NewPort("tcp", portStr)
	if err != nil {
		return nil, err
	}
	return nat.PortSet{
		port: struct{}{},
	}, nil
}

// startContainer starts a docker container and returns the container ID
// as well as a websocket connection to the attach endpoint.
func (h *handler) startContainer(ctrInfo containerInfo) (string, *websocket.Conn, error) {
	// pull alpine image if we don't already have it
	if err := h.pullImage(ctrInfo.dockerImage); err != nil {
		logrus.Fatalf("pulling %s failed: %v", ctrInfo.dockerImage, err)
	}

	securityOpts := []string{
		"no-new-privileges",
	}
	b := bytes.NewBuffer(nil)
	if err := json.Compact(b, []byte(seccompProfile)); err != nil {
		return "", nil, fmt.Errorf("compacting json for seccomp profile failed: %v", err)
	}
	securityOpts = append(securityOpts, fmt.Sprintf("seccomp=%s", b.Bytes()))

	dropCaps := &strslice.StrSlice{"NET_RAW"}

	port, err := generatePort(ctrInfo.port)
	if err != nil {
		return "", nil, err
	}

	// Add defaults to the docker image
	image := ctrInfo.dockerImage
	if image == "" {
		image = defaultDockerImage
	}

	// create the container
	r, err := h.dcli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        image,
			Cmd:          []string{"sh"},
			Tty:          true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
			StdinOnce:    true,
			ExposedPorts: port,
		},
		&container.HostConfig{
			SecurityOpt: securityOpts,
			CapDrop:     *dropCaps,
			NetworkMode: "none",
			LogConfig: container.LogConfig{
				Type: "none",
			},
			Resources: container.Resources{
				PidsLimit: 5,
			},
		},
		nil, "")
	if err != nil {
		return "", nil, err
	}

	// connect to the attach websocket endpoint
	header := http.Header(make(map[string][]string))
	header.Add("Origin", h.dockerURL.String())
	v := url.Values{
		"stdin":  []string{"1"},
		"stdout": []string{"1"},
		"stderr": []string{"1"},
		"stream": []string{"1"},
	}
	wsURL := fmt.Sprintf("ws://%s/%s/containers/%s/attach/ws?%s", h.dockerURL.Host, dockerAPIVersion, r.ID, v.Encode())
	var dialer = &websocket.Dialer{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: h.tlsConfig,
	}
	conn, _, err := dialer.Dial(wsURL, header)
	if err != nil {
		return r.ID, nil, fmt.Errorf("dialing %s with header %#v failed: %v", wsURL, header, err)
	}

	// start the container
	if err := h.dcli.ContainerStart(context.Background(), r.ID, types.ContainerStartOptions{}); err != nil {
		return r.ID, conn, err
	}

	return r.ID, conn, nil
}

// removeContainer removes with force a container by it's container ID.
func (h *handler) removeContainer(cid string) error {
	if err := h.dcli.ContainerRemove(context.Background(), cid,
		types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}); err != nil {
		return err
	}

	logrus.Debugf("removed container: %s", cid)

	return nil
}

// pullImage requests a docker image if it doesn't exist already.
func (h *handler) pullImage(image string) error {
	exists, err := h.imageExists(image)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	resp, err := h.dcli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	fd, isTerm := term.GetFdInfo(os.Stdout)

	return jsonmessage.DisplayJSONMessagesStream(resp, os.Stdout, fd, isTerm, nil)
}

// imageExists checks if a docker image exists.
func (h *handler) imageExists(image string) (bool, error) {
	_, _, err := h.dcli.ImageInspectWithRaw(context.Background(), image)
	if err == nil {
		return true, nil
	}

	if client.IsErrNotFound(err) {
		return false, nil
	}

	return false, err
}
