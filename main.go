package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/genuinetools/contained.af/version"
	"github.com/genuinetools/pkg/cli"
	"github.com/sirupsen/logrus"
)

const (
	defaultStaticDir        = "/usr/src/contained.af"
	defaultDockerHost       = "http://127.0.0.1:2375"
	defaultDockerUserNSHost = "http://127.0.0.1:2376"
	defaultDockerImage      = "alpine:latest"
)

var (
	dockerHost       string
	dockerUserNSHost string
	dockerCACert     string
	dockerCert       string
	dockerKey        string

	staticDir string
	port      string

	debug  bool
	tls_ws bool
)

func main() {
	var hostOS string
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "contained.af"
	p.Description = "A game for learning about containers, capabilities, and syscalls"

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.StringVar(&dockerHost, "dhost", defaultDockerHost, "host to commmunicate with docker on")
	p.FlagSet.StringVar(&dockerUserNSHost, "dusernshost", defaultDockerUserNSHost, "host to communicate with user namespace enabled docker on")
	p.FlagSet.StringVar(&dockerCACert, "dcacert", "", "trust certs signed only by this CA for docker host")
	p.FlagSet.StringVar(&dockerCert, "dcert", "", "path to TLS certificate file for docker host")
	p.FlagSet.StringVar(&dockerKey, "dkey", "", "path to TLS key file for docker host")
	p.FlagSet.StringVar(&hostOS, "os", "", "operating system of the docker host")

	p.FlagSet.StringVar(&staticDir, "frontend", defaultStaticDir, "directory that holds the static frontend files")
	p.FlagSet.StringVar(&port, "port", "10000", "port for server")

	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")
	p.FlagSet.BoolVar(&tls_ws, "tlsws", false, "enable TLS for container websocket")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		return nil
	}

	// Set the main program action.
	p.Action = func(ctx context.Context, args []string) error {
		if err := renderIndexPage(hostOS); err != nil {
			logrus.Fatal(err)
		}

		dockerURL, err := url.Parse(dockerHost)
		if err != nil {
			logrus.Fatalf("parsing docker daemon URL: %v", err)
		}

		dockerUserNSURL, err := url.Parse(dockerUserNSHost)
		if err != nil {
			logrus.Fatalf("parsing user namespace enabled docker daemon URL: %v", err)
		}

		// setup client TLS
		tlsConfig := tls.Config{
			// Prefer TLS1.2 as the client minimum
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
			InsecureSkipVerify: false,
		}

		if dockerCACert != "" {
			CAs, err := certPool(dockerCACert)
			if err != nil {
				logrus.Fatal(err)
			}
			tlsConfig.RootCAs = CAs
		}

		c := &http.Client{
			Transport: &http.Transport{},
		}
		if tls_ws {
			c = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tlsConfig,
				},
			}
		}

		if dockerCert != "" && dockerKey != "" {
			tlsCert, err := tls.LoadX509KeyPair(dockerCert, dockerKey)
			if err != nil {
				logrus.Fatalf("Could not load X509 key pair: %v. Make sure the key is not encrypted", err)
			}
			tlsConfig.Certificates = []tls.Certificate{tlsCert}
		}

		defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
		dcli, err := client.NewClient(dockerHost, "", c, defaultHeaders)
		if err != nil {
			logrus.Fatalf("creating docker client: %v", err)
		}

		dockerUserNSCLI, err := client.NewClient(dockerUserNSHost, "", c, defaultHeaders)
		if err != nil {
			logrus.Fatalf("creating user namespace enabled docker client: %v", err)
		}

		h := &handler{
			dcli:      dcli,
			dockerURL: dockerURL,
			tlsConfig: &tlsConfig,

			dUserNSCli:      dockerUserNSCLI,
			dockerUserNSURL: dockerUserNSURL,
			tls_ws:          tls_ws,
		}

		// ping handler
		http.HandleFunc("/ping", pingHandler)

		// info handler
		http.HandleFunc("/info", h.infoHandler)
		http.HandleFunc("/info-userns", h.infoUserNSHandler)

		// select profiles and websocket handling
		http.HandleFunc("/profiles", h.profilesHandler)

		// static files
		http.Handle("/", http.FileServer(http.Dir(staticDir)))

		logrus.Debugf("Server listening on %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			logrus.Fatalf("starting server failed: %v", err)
		}
		return nil
	}

	// Run our program.
	p.Run()
}

func renderIndexPage(hostOS string) error {
	tmplData := struct {
		OperatingSystem string
	}{
		OperatingSystem: hostOS,
	}

	tmpl, err := template.ParseFiles(filepath.Join(defaultStaticDir, "index-template.html"))
	if err != nil {
		// should be caught in development time and not fail at runtime
		return fmt.Errorf("static template parsing failed, %v", err)
	}

	indexFile, err := os.Create(filepath.Join(defaultStaticDir, "index.html"))
	if err != nil {
		return fmt.Errorf("could not create index.html, %v", err)
	}

	if err = tmpl.Execute(indexFile, tmplData); err != nil {
		return fmt.Errorf("executing template: %v", err)
	}
	return nil
}

// certPool returns an X.509 certificate pool from `caFile`, the certificate file.
func certPool(caFile string) (*x509.CertPool, error) {
	// If we should verify the server, we need to load a trusted ca
	certPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("could not read CA certificate %q: %v", caFile, err)
	}
	if !certPool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("failed to append certificates from PEM file: %q", caFile)
	}
	s := certPool.Subjects()
	subjects := make([]string, len(s))
	for i, subject := range s {
		subjects[i] = string(subject)
	}
	logrus.Debugf("Trusting certs with subjects: %v", subjects)
	return certPool, nil
}
