package service

import (
	"context"
	"errors"
	"microservicePeoject/discovery/config"
	"microservicePeoject/discovery/discover"
)

type Service interface {
	// HealthCheck 健康检查接口
	HealthCheck() bool

	// SayHello 打招呼接口
	SayHello() string

	//DiscoveryService 服务发现接口
	DiscoveryService(ctx context.Context, serviceName string) ([]interface{}, error)
}

var ErrNotServiceInstances = errors.New("instances are not exists")

// DiscoveryServiceImpl 具体的实现接口Service的结构体
type DiscoveryServiceImpl struct {
	discoveryClient discover.DiscoveryClient
}

func NewDiscoveryServiceImpl(discoveryClient discover.DiscoveryClient) Service {
	return &DiscoveryServiceImpl{
		discoveryClient: discoveryClient,
	}
}

func (*DiscoveryServiceImpl) HealthCheck() bool {
	return true
}

func (*DiscoveryServiceImpl) SayHello() string {
	return "hello world"
}

func (service *DiscoveryServiceImpl) DiscoveryService(ctx context.Context, serviceName string) ([]interface{}, error) {
	instances := service.discoveryClient.DiscoveryServices(serviceName, config.Logger)
	if instances == nil || len(instances) == 0 {
		return nil, ErrNotServiceInstances
	}
	return instances, nil
}
