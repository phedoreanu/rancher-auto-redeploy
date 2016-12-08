# rancher-auto-redeploy 
[![Build Status](https://travis-ci.org/phedoreanu/rancher-auto-redeploy.svg?branch=master)](https://travis-ci.org/phedoreanu/rancher-auto-redeploy) [![Coverage Status](https://coveralls.io/repos/github/phedoreanu/rancher-auto-redeploy/badge.svg?branch=master)](https://coveralls.io/github/phedoreanu/rancher-auto-redeploy?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/phedoreanu/rancher-auto-redeploy)](https://goreportcard.com/report/github.com/phedoreanu/rancher-auto-redeploy) [![GoDoc](https://godoc.org/github.com/phedoreanu/rancher-auto-redeploy?status.svg)](https://godoc.org/github.com/phedoreanu/rancher-auto-redeploy)  
Go HTTP callback Rancher auto-redeploy server.

## How it works
`rancher-auto-redeploy` listens for HTTP requests from Docker Hub Webhooks on a configurable address and port, secured using a random secret. Whenever there is a notification of a new image being pushed, it scans all services in the Rancher environment and fires an upgrade for all of those that match the published image and tag.

## Usage
The easiest way to run `rancher-auto-redeploy` on your Rancher server is through its [docker image.](https://hub.docker.com/r/phedoreanu/rancher-auto-redeploy/) 
Run a new stack using the manastech/rancher-autoredeploy image, and configure it with the required credentials.
```yaml
redeployer:
  ports:
  - 8091:8091
  environment:
    DOCKER_HUB_KEY: randomsecret
    RANCHER_API_KEY: your.api.key
    RANCHER_API_SECRET: your.api.secret
    RANCHER_HOST: your.racher.host
    RANCHER_PROJECT_ID: your.project.id
  labels:
    io.rancher.container.pull_image: always
  image: phedoreanu/rancher-auto-redeploy:latest
```
This will start an instance of the redeployer image, listening on port 8090 for requests from docker hub including the `randomsecret` key, and updating the services in project `yourprojectid` at `your.rancher.server`. Note that the bind port and address can be configured to the values that best match your setup.

The next step is to add a webhook to your docker hub repositories, so they will notify the `redeployer` container when there is a new version of an image. On the webhooks tab of your project, simply add an entry to `http://your.rancher.node:8091/randomsecret`.

*Important:*
 * set `your.racher.host` without `/v1/...` i.e. `https://rancher.example.com`.
 * remember to set `io.rancher.container.pull_image` label on your service to `always`, to ensure that the new image is actually pulled when upgrading. Future versions of this project may set this label automatically.
