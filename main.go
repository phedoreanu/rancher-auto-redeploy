package main

import (
	"flag"
	"os"

	"github.com/phedoreanu/rancher-auto-redeploy/api"
)

func main() {
	BindPort := flag.Int("port", 8091, "Port to bind for listening for Docker Hub Webhooks.")
	BindAddress := flag.String("address", "0.0.0.0", "Bind Address to bind for listening for Docker Hub webhooks.")

	redeploy := &api.RancherAPI{
		Address:      *BindAddress,
		Port:         *BindPort,
		Host:         os.Getenv("RANCHER_HOST"),
		ProjectID:    os.Getenv("RANCHER_PROJECT_ID"),
		APIKey:       os.Getenv("RANCHER_API_KEY"),
		APISecret:    os.Getenv("RANCHER_API_SECRET"),
		DockerHubKey: os.Getenv("DOCKER_HUB_KEY"),
	}
	redeploy.Run()
}
