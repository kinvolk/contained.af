# Container Escape Bounty CTF version of contained.af

This is a fork of [contained.af](https://github.com/genuinetools/contained.af)
(a [game](https://contained.af/) for learning about containers, capabilities, and syscalls).
It is used in the [Container Escape Bounty CTF](https://github.com/kinvolk/container-escape-bounty)
which defines container profiles and creates VMs.
On these VMs a researcher can spawn containers through a web interface
and has to break out.

## Usage
This is intended to be used from the terraform files in other repository linked above.
The defined profiles can only be tested on real VMs.

## Local Test/Development Mode

To spare the time creating a cluster, there are makefiles inherited from contained.af
that use a Docker-in-Docker setup.

These are the components involved:

  * A static HTML and JavaScript frontend in `frontend/`
  * A Go web server in the project root
  * An isolated Docker installation, running inside a Docker container
    ("Docker-in-Docker").

Start an isolated Docker instance in the background with:

```
make dind
```

Build and run the server with:

```
make run
```

To show the button to disable SELinux or AppArmor, you need to provide
either `fedora` or `ubuntu` in the OS variable:

```
make run ARGS="-os fedora"
```

After a few moments, contained will be available at http://localhost:10000/.
