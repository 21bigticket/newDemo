package config

import (
	"strings"
	"sync"

	conf "dubbo.apache.org/dubbo-go/v3/common/config"
	"dubbo.apache.org/dubbo-go/v3/config_center"
	"github.com/dubbogo/gost/log/logger"
	"gopkg.in/yaml.v3"
)

// AppConfig 应用配置（全局单例）
var appConfig = &AppConfigManager{
	data: make(map[string]interface{}),
	mu:   sync.RWMutex{},
}

// AppConfigManager 应用配置管理器
type AppConfigManager struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// Init 从 dubbo-go 配置中心初始化应用配置
// 必须在 dubbo_config.Load() 之后调用
func InitAppConfig(dataID, group string) error {
	dynamicConfig := conf.GetEnvInstance().GetDynamicConfiguration()
	if dynamicConfig == nil {
		return nil // 配置中心未启动，返回nil
	}

	// 获取配置内容
	content, err := dynamicConfig.GetProperties(dataID, config_center.WithGroup(group))
	if err != nil {
		logger.Warnf("Failed to get config from center: %v", err)
		return err
	}

	// 解析配置
	var configMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &configMap); err != nil {
		logger.Errorf("Failed to parse config: %v", err)
		return err
	}

	// 更新配置
	appConfig.mu.Lock()
	appConfig.data = configMap
	appConfig.mu.Unlock()

	logger.Infof("App config initialized: %+v", configMap)

	// 添加配置监听器
	dynamicConfig.AddListener(dataID, &appConfigListener{}, config_center.WithGroup(group))

	return nil
}

// appConfigListener 配置监听器
type appConfigListener struct{}

// Process 实现 ConfigurationListener 接口
func (l *appConfigListener) Process(event *config_center.ConfigChangeEvent) {
	logger.Infof("App config changed: %v", event.Value)

	valueStr, ok := event.Value.(string)
	if !ok {
		logger.Errorf("Failed to convert config value to string")
		return
	}

	var configMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(valueStr), &configMap); err != nil {
		logger.Errorf("Failed to parse config: %v", err)
		return
	}

	// 更新配置
	appConfig.mu.Lock()
	appConfig.data = configMap
	appConfig.mu.Unlock()

	logger.Infof("App config updated successfully")
}

// Get 获取配置值（支持点号路径，如 "redis.host"）
func Get(key string) interface{} {
	appConfig.mu.RLock()
	defer appConfig.mu.RUnlock()

	// 支持点号路径，如 "redis.host"
	keys := strings.Split(key, ".")
	var current interface{} = appConfig.data

	for _, k := range keys {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[k]
		} else {
			return nil
		}
	}

	return current
}

// GetString 获取字符串配置
func GetString(key string) string {
	val := Get(key)
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetInt 获取整数配置
func GetInt(key string) int {
	val := Get(key)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

// GetBool 获取布尔配置
func GetBool(key string) bool {
	val := Get(key)
	if val == nil {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

// GetStringMap 获取map配置
func GetStringMap(key string) map[string]interface{} {
	val := Get(key)
	if val == nil {
		return nil
	}
	if m, ok := val.(map[string]interface{}); ok {
		return m
	}
	return nil
}

// IsSet 检查配置是否存在
func IsSet(key string) bool {
	return Get(key) != nil
}

// GetAll 获取所有配置
func GetAll() map[string]interface{} {
	appConfig.mu.RLock()
	defer appConfig.mu.RUnlock()

	// 返回副本
	result := make(map[string]interface{})
	for k, v := range appConfig.data {
		result[k] = v
	}
	return result
}
