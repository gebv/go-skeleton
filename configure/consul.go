package configure

import (
	"net"
	"strconv"

	consul "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
)

// RegisterService registration of service in the Consul.
func RegisterService(consulClient *consul.Client, service, metricsHostPort string) error {
	_, metricsPortS, _ := net.SplitHostPort(metricsHostPort)
	metricsPort, err := strconv.Atoi(metricsPortS)
	if err != nil {
		return errors.Wrap(err, "invalid host port of metrics")
	}

	return consulClient.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		Name:    service,
		Address: serviceAddress,
		Port:    metricsPort,
	})
}
