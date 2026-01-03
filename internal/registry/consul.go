package registry

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hashicorp/consul/api"
)

var client *api.Client

func InitConsul() {
	ctx := context.Background()
	cfg := api.DefaultConfig()
	addr := g.Cfg().MustGet(ctx, "consul.addr").String()
	cfg.Address = addr

	var err error
	client, err = api.NewClient(cfg)
	if err != nil {
		g.Log().Fatalf(ctx, "init consul failed: %v", err)
	}
}

func RegisterService() error {
	ctx := context.Background()
	serviceName := g.Cfg().MustGet(ctx, "consul.serviceName").String()
	servicePort := g.Cfg().MustGet(ctx, "consul.servicePort").Int()
	serviceAddr := g.Cfg().MustGet(ctx, "consul.serviceAddr").String()

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s", serviceName, serviceAddr),
		Name:    serviceName,
		Address: serviceAddr,
		Port:    servicePort,
		Check: &api.AgentServiceCheck{
			TCP:                            fmt.Sprintf("%s:%d", serviceAddr, servicePort),
			Interval:                       "15s",
			Timeout:                        "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("register service failed: %v", err)
	}

	g.Log().Infof(ctx, "service registered: %s at %s:%d", serviceName, serviceAddr, servicePort)
	return nil
}

func DeregisterService() error {
	ctx := context.Background()
	serviceID := "jh_admin_service-1"

	if err := client.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("deregister service failed: %v", err)
	}

	g.Log().Infof(ctx, "service deregistered: jh_admin_service")
	return nil
}
