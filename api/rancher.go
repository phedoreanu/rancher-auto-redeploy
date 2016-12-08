package api

// Response mimics Rancher's Services JSON response.
type Response struct {
	Data []*Service
}

// LaunchConfig mimics Rancher's LaunchConfig JSON response.
type LaunchConfig struct {
	ImageUUID   string            `json:"imageUuid"`
	Kind        string            `json:"kind"`
	Ports       []string          `json:"ports"`
	Labels      map[string]string `json:"labels"`
	Environment map[string]string `json:"environment"`
}

// Service mimics Rancher's Service JSON response.
type Service struct {
	ID, Kind, Type, Name, State, AccountID, FQDN string
	LaunchConfig                                 *LaunchConfig
	Actions                                      map[string]string
}

// Strategy mimics Rancher's Strategy JSON response.
type Strategy struct {
	InServiceStrategy *InServiceStrategy `json:"inServiceStrategy"`
	ToServiceStrategy *InServiceStrategy `json:"toServiceStrategy"`
}

// InServiceStrategy mimics Rancher's InServiceStrategy JSON response.
type InServiceStrategy struct {
	LaunchConfig *LaunchConfig `json:"launchConfig"`
}

// UpgradeResponse mimics Rancher's UpgradeResponse JSON response.
type UpgradeResponse struct {
	ID, Type, Code, Message, FieldName, State string
	Status                                    int
	Actions                                   map[string]string
}
