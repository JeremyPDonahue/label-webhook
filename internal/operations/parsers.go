package operations

import (
	"encoding/json"

	dep "k8s.io/api/apps/v1"
	pod "k8s.io/api/core/v1"
)

func parseDeployment(object []byte) (*dep.Deployment, error) {
	var dp dep.Deployment
	if err := json.Unmarshal(object, &dp); err != nil {
		return nil, err
	}

	return &dp, nil
}

func parsePod(object []byte) (*pod.Pod, error) {
	var pod pod.Pod
	if err := json.Unmarshal(object, &pod); err != nil {
		return nil, err
	}

	return &pod, nil
}
