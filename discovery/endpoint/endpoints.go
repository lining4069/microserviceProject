package endpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"microservicePeoject/discovery/service"
)

/**
与 transport 层交互
请求操作点
负责将transport传入的解码请求转发给Service的对应处理方法。

生命端点，并提供端点的构建方法以及数据交互数据格式。

*/

type DiscoveryEndpoints struct {
	SayHelloEndpoint    endpoint.Endpoint
	DiscoveryEndpoint   endpoint.Endpoint
	HealthCheckEndpoint endpoint.Endpoint
}

// SayHelloRequest 打招呼请求结构体
type SayHelloRequest struct {
}

// SayHelloResponse 打招呼响结构体
type SayHelloResponse struct {
	Message string `json:"message"`
}

func MakeSayHelloEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		message := svc.SayHello()
		return SayHelloResponse{
			Message: message,
		}, nil

	}
}

type DiscoveryRequest struct {
	ServiceName string
}

type DiscoveryResponse struct {
	Instances []interface{} `json:"instances"`
	Error     string        `json:"error"`
}

func MakeDiscoveryEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(DiscoveryRequest)
		instances, err := svc.DiscoveryService(ctx, req.ServiceName)
		var errString = ""
		if err != nil {
			errString = err.Error()
		}
		return &DiscoveryResponse{
			Instances: instances,
			Error:     errString,
		}, nil
	}
}

type HealthRequest struct {
}

type HealthResponse struct {
	Status bool `json:"Status"`
}

func MakeHealthCheckEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return HealthResponse{
			Status: status,
		}, nil
	}
}
