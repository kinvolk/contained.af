package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	dockerAPIVersion = "v1.35"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type handler struct {
	// parameters for normal docker daemon
	dcli      *client.Client
	dockerURL *url.URL

	// parameters for docker daemon with user namespace enabled
	dUserNSCli      *client.Client
	dockerUserNSURL *url.URL

	tlsConfig *tls.Config
	tls_ws     bool
}

func (h *handler) client(userns bool) *client.Client {
	if userns {
		return h.dUserNSCli
	}
	return h.dcli
}
func (h *handler) url(userns bool) *url.URL {
	if userns {
		return h.dockerUserNSURL
	}
	return h.dockerURL
}

type message struct {
	Type   string `json:"type"`
	Data   string `json:"data"`
	Height uint   `json:"height,omitempty"`
	Width  uint   `json:"width,omitempty"`
}

// pingHander returns pong.
func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "pong")
}

func constructContainerInfo(r *http.Request) (*containerInfo, error) {
	var c containerInfo
	if len(r.URL.Query()["port"]) > 0 {
		c.port = r.URL.Query()["port"][0]
	}

	if len(r.URL.Query()["image"]) > 0 {
		c.dockerImage = r.URL.Query()["image"][0]
	}

	if len(r.URL.Query()["profile"]) > 0 {
		c.dockerProfile = dockerProfile(r.URL.Query()["profile"][0])

		if _, ok := dockerProfiles[c.dockerProfile]; !ok {
			return nil, fmt.Errorf("Docker profile %q is invalid.", c.dockerProfile)
		}
	}

	if len(r.URL.Query()["userns"]) > 0 {
		val := r.URL.Query()["userns"][0]
		if val == "enabled" {
			c.userns = true
		}
	}

	if len(r.URL.Query()["selinux"]) > 0 {
		val := r.URL.Query()["selinux"][0]
		// By default all the containers start with selinux enabled. Now user
		// can override this default behavior.
		c.selinux = true
		if val == "disabled" {
			c.selinux = false
		}
	}

	if len(r.URL.Query()["apparmor"]) > 0 {
		val := r.URL.Query()["apparmor"][0]
		// By default all the containers start with apparmor enabled. Now user
		// can override this default behavior.
		c.apparmor = true
		if val == "disabled" {
			c.apparmor = false
		}
	}

	return &c, nil
}

func (h *handler) profilesHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("websocket upgrader failed: %v", err)
		return
	}

	ctrInfo, err := constructContainerInfo(r)
	if err != nil {
		logrus.Errorf("generating container info failed: %v", err)
		data := message{
			Type: "stdout",
			Data: fmt.Sprintf("generating container info failed: %v", err),
		}
		if err := conn.WriteJSON(data); err != nil {
			logrus.Errorf("writing error message to browser websocket failed: %v", err)
		}
		return
	}

	// start the container and create the container websocket connection
	containerWSConn, err := h.startContainer(ctrInfo)
	if err != nil {
		logrus.Errorf("starting container failed: %v", err)
		data := message{
			Type: "stdout",
			Data: fmt.Sprintf("starting container failed: %v", err),
		}
		if err := conn.WriteJSON(data); err != nil {
			logrus.Errorf("writing error message to browser websocket failed: %v", err)
		}
		return
	}
	defer containerWSConn.Close()
	logrus.Infof("container started with id: %s", ctrInfo.containerid)

	// start a go routine to listen on the container websocket and send to the browser websocket
	done := make(chan struct{})
	go func() {
		defer containerWSConn.Close()
		defer close(done)

		for {
			// TODO: this will panic if we couldn't start the container
			_, msg, err := containerWSConn.ReadMessage()
			if err != nil {
				if e, ok := err.(*websocket.CloseError); ok {
					logrus.Warnf("container websocket closed %s %d", e.Text, e.Code)
					// cleanup and remove the container
					if err := h.removeContainer(ctrInfo); err != nil {
						logrus.Errorf("removing container %s failed: %v", ctrInfo.containerid, err)
					}
					// cleanly close the browser connection
					if err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
						logrus.Errorf("closing browser websocket failed: %v", err)
					}
					break
				}
				if err == io.EOF {
					continue
				}
				logrus.Errorf("reading from container websocket failed: %v", err)
				continue
			}
			logrus.Debugf("received from container websocket: %s", string(msg))

			// send it back through to the browser websocket as a binary frame
			b := message{
				Type: "stdout",
				Data: string(msg),
			}
			if err := conn.WriteJSON(b); err != nil {
				if err == websocket.ErrCloseSent {
					logrus.Warn("browser websocket close sent")
					// cleanup and remove the container
					if err := h.removeContainer(ctrInfo); err != nil {
						logrus.Errorf("removing container %s failed: %v", ctrInfo.containerid, err)
					}
					break
				}
				logrus.Errorf("writing to browser websocket failed: %v", err)
				continue
			}
			logrus.Debugf("wrote to browser websocket: %#v", b)
		}
	}()

	for {
		var data message
		if err := conn.ReadJSON(&data); err != nil {
			if e, ok := err.(*websocket.CloseError); ok {
				logrus.Warnf("browser websocket closed %s %d", e.Text, e.Code)
				// cleanup and remove the container
				if err := h.removeContainer(ctrInfo); err != nil {
					logrus.Errorf("removing container %s failed: %v", ctrInfo.containerid, err)
				}
				break
			}
			logrus.Errorf("reading from browser websocket failed: %v", err)
			continue
		}
		logrus.Debugf("recieved from browser websocket: %#v", data)

		// send to container websocket or resize
		switch data.Type {
		case "stdin":
			if len(data.Data) > 0 {
				if err := containerWSConn.WriteMessage(websocket.TextMessage, []byte(data.Data)); err != nil {
					if err == websocket.ErrCloseSent {
						logrus.Warn("container websocket close sent")
						// cleanup and remove the container
						if err := h.removeContainer(ctrInfo); err != nil {
							logrus.Errorf("removing container %s failed: %v", ctrInfo.containerid, err)
						}
						break
					}
					logrus.Errorf("writing to container websocket failed: %v", err)
					continue
				}
				logrus.Debugf("wrote to container websocket: %q", data.Data)
			}
		case "resize":
			if err := h.dcli.ContainerResize(context.Background(), ctrInfo.containerid, types.ResizeOptions{
				Height: data.Height,
				Width:  data.Width,
			}); err != nil {
				logrus.Errorf("resize container to height -> %d, width: %d failed: %v", data.Height, data.Width, err)
			}
		default:
			logrus.Warnf("got unknown data type: %s", data.Type)
		}
	}

	// cleanup and remove the container
	if err := h.removeContainer(ctrInfo); err != nil {
		logrus.Errorf("removing container %s failed: %v", ctrInfo.containerid, err)
	}
}

// infoHander returns information about the connected docker daemon.
func (h *handler) infoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := retrieveInfo(w, r, h.dcli); err != nil {
		logrus.Errorf("docker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// infoUserNSHandler returns information about the connected docker daemon that
// supports user namespace
func (h *handler) infoUserNSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := retrieveInfo(w, r, h.dUserNSCli); err != nil {
		logrus.Errorf("docker user namespace: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func retrieveInfo(w http.ResponseWriter, r *http.Request, client *client.Client) error {

	info, err := client.Info(context.Background())
	if err != nil {
		return fmt.Errorf("getting docker info failed: %v", err)
	}

	b, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal indent info failed: %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", b)
	return nil
}
