package instance

import (
	"helloworld/pkg/config"

	"dubbo.apache.org/dubbo-go/v3"
	"dubbo.apache.org/dubbo-go/v3/config_center"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/registry"

	"github.com/dubbogo/gost/log/logger"
)

func InitInstance(cfg *config.Config) (*dubbo.Instance, error) {
	ins, err := dubbo.NewInstance(
		dubbo.WithName(cfg.AppName),
		dubbo.WithConfigCenter(
			config_center.WithNacos(),
			config_center.WithDataID(cfg.AppName),
			config_center.WithAddress(cfg.Nacos.Address),
			config_center.WithNamespace(cfg.Nacos.Namespace),
			config_center.WithGroup(cfg.Nacos.Group),
		),
		dubbo.WithRegistry(
			registry.WithNacos(),
			registry.WithAddress(cfg.Nacos.Address),
		),
		dubbo.WithProtocol(
			protocol.WithTriple(),
			protocol.WithPort(cfg.AppPort),
		),
	)
	if err != nil {
		logger.Errorf("new dubbo instance failed: %v", err)
		panic(err)
	}
	return ins, err
}
