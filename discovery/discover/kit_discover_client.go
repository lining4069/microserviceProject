package discover

import (
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"log"
	"strconv"
	"sync"
)

type KitDiscoverClient struct {
	Host   string
	Port   int
	client consul.Client
	// 增加缓存机制，避免每次服务发现，都与consul交互一次
	// 将服务实例信息列表按照服务名的方式组织缓存到服务注册与发现客户端本地，并通过Consul提供的Watch机制监控该服务名小下服务实例数据的变化

	config       *api.Config
	mutex        sync.Mutex
	instancesMap sync.Map // 服务实例缓存字段
}

func NewKitDiscoverClient(consulHost string, consulPort int) (DiscoveryClient, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulHost + ":" + strconv.Itoa(consulPort)

	apiClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}

	client := consul.NewClient(apiClient)
	return &KitDiscoverClient{
		Host:   consulHost,
		Port:   consulPort,
		client: client,
		config: consulConfig,
		//不需要显式初始化 mutex 字段，因为 sync.Mutex 类型的零值就是已经初始化的互斥锁，可以直接使用。
	}, nil

}

func (consulClient *KitDiscoverClient) Register(serviceName, instanceId, healthCheckUrl string, instanceHost string, instancePort int, meta map[string]string, logger *log.Logger) bool {
	// 使用consul.Client 的Register方法
	serviceRegisteration := &api.AgentServiceRegistration{
		ID:      instanceId,
		Name:    serviceName,
		Address: instanceHost,
		Port:    instancePort,
		Meta:    meta,
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           "http://" + instanceHost + ":" + strconv.Itoa(instancePort) + healthCheckUrl,
			Interval:                       "15s",
		},
	}

	err := consulClient.client.Register(serviceRegisteration)
	if err != nil {
		log.Println("Register Service Error !")
		return false
	} else {
		log.Println("Register Service Success !")
	}
	return true
}

func (consulClient *KitDiscoverClient) DeRegister(instanceId string, logger *log.Logger) bool {
	serviceRegisteration := &api.AgentServiceRegistration{
		ID: instanceId,
	}

	err := consulClient.client.Deregister(serviceRegisteration)
	if err != nil {
		log.Println("DeRegister Service Error !")
		return false
	} else {
		log.Println("DeRegister Service Success !")
	}
	return true
}

func (consulClient *KitDiscoverClient) DiscoveryServices(serviceName string, logger *log.Logger) []interface{} {

	// 该服务已经缓存
	instanceList, ok := consulClient.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]interface{})
	}
	// 申请锁
	consulClient.mutex.Lock()
	// 再次检查是否被监控缓存
	instanceList, ok = consulClient.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]interface{})
	} else {
		// 注册监控
		go func() {
			//使用consul服务实例监控来监控某个服务的服务实例列表变化
			params := make(map[string]interface{})
			params["type"] = "service"
			params["service"] = serviceName
			plan, _ := watch.Parse(params)
			plan.Handler = func(u uint64, i interface{}) {
				if i == nil {
					return
				}
				v, ok := i.([]*api.ServiceEntry)
				if !ok {
					return // 数据异常，忽略
				}
				// 没有服务实例在线
				if len(v) == 0 {
					consulClient.instancesMap.Store(serviceName, []interface{}{})
				}
				var healthServices []interface{}
				for _, service := range v {
					if service.Checks.AggregatedStatus() == api.HealthPassing {
						healthServices = append(healthServices, service.Service)
					}
				}
				consulClient.instancesMap.Store(serviceName, healthServices)
			}
			defer plan.Stop()
			plan.Run(consulClient.config.Address)
		}()
	}
	// 根据服务名称获取服务实例列表
	entries, _, err := consulClient.client.Service(serviceName, "", false, nil)
	if err != nil {
		log.Println("Discover Service Error !")
		return nil
	}

	instances := make([]interface{}, len(entries))
	for i := 0; i < len(instances); i++ {
		instances[i] = entries[i].Service
	}
	consulClient.instancesMap.Store(serviceName, instances)
	return instances
}
