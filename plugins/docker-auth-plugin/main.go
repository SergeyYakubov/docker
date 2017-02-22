package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/authorization"
	"os/user"
	"strconv"
)

const (
	//defaultDockerHost   = "unix:///var/run/docker.sock"
	defaultDockerHost   = "tcp://localhost:2376"
	defaultPluginSocket = "/var/run/docker/plugins/docker-auth-plugin.sock"
)

var (
	flDockerHost   = flag.String("host", defaultDockerHost, "Docker host the plugin connects to when inspecting")
	flPluginSocket = flag.String("socket", defaultPluginSocket, "Plugin's socket path")
)

func main() {
	flag.Parse()

	novolume, err := newPlugin(*flDockerHost)
	if err != nil {
		logrus.Fatal(err)
	}
	u, _ := user.Lookup("root")
	gid, _ := strconv.Atoi(u.Gid)

	h := authorization.NewHandler(novolume)
	if err := h.ServeUnix(*flPluginSocket, gid); err != nil {
		logrus.Fatal(err)
	}
}
