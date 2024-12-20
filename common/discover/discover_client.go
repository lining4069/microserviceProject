package discover

import "log"

type DiscoveryClient interface {
	/*
		服务注册接口
		serviceName 服务名
		instanceId 服务实例id
		instancePort 服务实例端口
		healthCheckUrl 健康检查接口
		instanceHost 服务器实例地址
		meta 服务实例元数据

		用于服务注册，服务实例将自身所述服务名和服务元数据注册到Consul中
	*/
	Register(
		serviceName, instanceId, healthCheckUrl string,
		instanceHost string, instancePort int,
		meta map[string]string, logger *log.Logger,
	) bool

	/*
		服务注销接口
		instanceId 服务器实例id

		用于服务注销，服务关闭时 请求Consul将自身元数据注销，避免无效请求
	*/
	DeRegister(instanceId string, logger *log.Logger) bool

	/*
		服务发现接口
		serviceName 服务名

		用于服务发现，通过服务名向Consul请求对应的服务实例信息列表。
	*/
	DiscoveryServices(serviceName string, logger *log.Logger) []interface{}
}
