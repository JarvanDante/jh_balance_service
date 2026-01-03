package registry

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hashicorp/consul/api"
)

var client *api.Client

func InitConsul() error {
	ctx := context.Background()
	cfg := api.DefaultConfig()
	addr := g.Cfg().MustGet(ctx, "consul.addr").String()
	if addr == "" {
		return fmt.Errorf("consul.addr 配置不能为空")
	}
	cfg.Address = addr

	var err error
	client, err = api.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("init consul failed: %v", err)
	}

	// 自动注册服务
	return RegisterService()
}

func RegisterService() error {
	ctx := context.Background()
	serviceName := g.Cfg().MustGet(ctx, "consul.serviceName", "balance_service").String()
	servicePort := g.Cfg().MustGet(ctx, "consul.servicePort", 8000).Int()
	serviceAddr := g.Cfg().MustGet(ctx, "consul.serviceAddr", "127.0.0.1").String()

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
	serviceName := g.Cfg().MustGet(ctx, "consul.serviceName", "balance_service").String()
	serviceAddr := g.Cfg().MustGet(ctx, "consul.serviceAddr", "127.0.0.1").String()
	serviceID := fmt.Sprintf("%s-%s", serviceName, serviceAddr)

	if err := client.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("deregister service failed: %v", err)
	}

	g.Log().Infof(ctx, "service deregistered: %s", serviceName)
	return nil
}
