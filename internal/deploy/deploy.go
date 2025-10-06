package deploy

import (
	"kubevirt.io/client-go/kubecli"
)

type DeployService struct {
	client kubecli.KubevirtClient
}

func NewDeployService(client kubecli.KubevirtClient) *DeployService {
	return &DeployService{
		client: client,
	}
}
