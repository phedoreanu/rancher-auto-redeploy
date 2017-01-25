package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

// RancherAPI is the main struct
type RancherAPI struct {
	Address, Host, APIKey, APISecret, ProjectID, DockerHubKey string
	Port                                                      int
	Services                                                  map[string]*Service
	LoadBalancers                                             map[string]*Service
}

func (ra *RancherAPI) rootHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintln(w, "Welcome to the Rancher auto-deployer!")
}

func (ra *RancherAPI) redeployHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	if key == "" {
		http.Error(w, "Received request without 'key' query parameter.", http.StatusUnauthorized)
		return
	} else if key != ra.DockerHubKey {
		http.Error(w, "Received request with invalid key.", http.StatusForbidden)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Received request with invalid content type.", http.StatusNotImplemented)
		return
	}

	var data DockerHubRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}
	image := fmt.Sprintf("%s:%s", data.Repository.RepoName, data.PushData.Tag)
	log.Printf("Received request for %s", image)

	s, ok := ra.Services[data.Repository.RepoName]
	if ok {
		ra.RefreshService(s)

		// setting new image & tag from DockerHub payload
		s.LaunchConfig.ImageUUID = "docker:" + image
	} else {
		ra.LoadServices()
	}
	ra.UpgradeService(s)

	go func() {
		time.Sleep(2 * time.Minute)
		for i := 0; i < 5; i++ {
			ra.RefreshService(s)

			if s.State == "upgraded" {
				if ra.FinishUpgradeService(s) {
					break
				}
			} else if s.State == "active" {
				log.Printf("Service %s active", s.ID)
				break
			} else {
				log.Printf("Service %s still upgrading. Retrying in 30 seconds...", s.ID)
				time.Sleep(30 * time.Second)
			}
		}
		time.Sleep(3 * time.Second)
		ra.RefreshService(s)
	}()

	payload := []byte(`{}`)
	req, err := http.NewRequest("POST", data.CallbackURL, bytes.NewReader(payload))
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}
}

// UpgradeService calls Rancher's JSON API with the `upgrade` action.
func (ra *RancherAPI) UpgradeService(s *Service) {
	if s.State == "upgraded" {
		log.Printf("Service %s upgraded, finishing upgrade", s.ID)
		ra.FinishUpgradeService(s)
		return
	}
	if s.State != "active" {
		log.Printf("Service %s not in active state, canceling upgrade", s.ID)
		return
	}

	strategy := Strategy{InServiceStrategy: &InServiceStrategy{LaunchConfig: s.LaunchConfig}}
	payload, _ := json.Marshal(strategy)
	req, err := http.NewRequest("POST", s.Actions["upgrade"], bytes.NewReader(payload))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(ra.APIKey, ra.APISecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	var upgrade UpgradeResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&upgrade)
	if err != nil {
		log.Println(err)
	}
	if upgrade.State == "upgrading" {
		log.Printf("Upgrading %s", s.ID)
	} else {
		log.Println(upgrade)
	}
}

// FinishUpgradeService calls Rancher's JSON API with the `finishupgrade` action.
func (ra *RancherAPI) FinishUpgradeService(s *Service) bool {
	log.Printf("Finishing %s (%s) upgrade", s.Name, s.ID)
	payload := []byte(`{}`)
	req, err := http.NewRequest("POST", s.Actions["finishupgrade"], bytes.NewReader(payload))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(ra.APIKey, ra.APISecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()

	var upgrade UpgradeResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&upgrade)
	if err != nil {
		log.Println(err)
	}
	if upgrade.State == "finishing-upgrade" {
		log.Printf("Successfully upgraded %s", s.ID)
		return true
	}
	log.Println(upgrade)
	return false
}

// RefreshService does a GET to Rancher's JSON API for that specific `Service`
func (ra *RancherAPI) RefreshService(s *Service) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v2-beta/projects/%s/services/%s", ra.Host, ra.ProjectID, s.ID), nil)
	req.SetBasicAuth(ra.APIKey, ra.APISecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	var service Service
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&service)
	if err != nil {
		log.Println(err)
	}
	if s.State != service.State {
		log.Printf("Service %s changed state from `%s` to `%s`", s.ID, s.State, service.State)
		s.State = service.State
		s.Actions = service.Actions
	}
}

// LoadServices calls Rancher's JSON API for all the services.
func (ra *RancherAPI) LoadServices() {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v2-beta/projects/%s/services", ra.Host, ra.ProjectID), nil)
	req.SetBasicAuth(ra.APIKey, ra.APISecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err.Error())
	}
	defer resp.Body.Close()

	var services Response
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&services)
	if err != nil {
		log.Panic(err)
	}

	ra.Services = make(map[string]*Service)
	ra.LoadBalancers = make(map[string]*Service)
	for _, s := range services.Data {
		switch serviceType := s.Type; serviceType {
		case "service":
			ra.Services[splitUUID(s.LaunchConfig.ImageUUID)] = s

			// check for `upgraded` services and finishing the upgrade
			if s.State == "upgraded" {
				ra.FinishUpgradeService(s)
			}
		case "loadBalancerService":
			ra.LoadBalancers[s.FQDN] = s
		}
	}
	log.Printf("Loaded %d services", len(ra.Services))
	log.Printf("Loaded %d loadBalancers", len(ra.LoadBalancers))
}

func splitUUID(uuid string) string {
	//"imageUuid": "docker:rancher/convoy-agent:v0.12.0",
	split := strings.Split(uuid, ":")
	return split[1]
}

// Run starts the callback HTTP server and listens on `BindAddress:BindPort`
func (ra *RancherAPI) Run() {
	go func() {
		time.Sleep(20 * time.Second)
		ra.LoadServices()
	}()

	router := httprouter.New()
	router.GET("/", ra.rootHandler)
	router.POST("/:key", ra.redeployHandler)
	bind := fmt.Sprintf("%s:%d", ra.Address, ra.Port)

	log.Printf("Listening for DockerHub webhooks on %s\n", bind)
	log.Fatal(http.ListenAndServe(bind, router))
}
