package config

import (
	"fmt"
	"strings"
	"sync"

	conf "dubbo.apache.org/dubbo-go/v3/common/config"
	"dubbo.apache.org/dubbo-go/v3/config_center"
	"github.com/dubbogo/gost/log/logger"
	"github.com/spf13/viper"
)

// ViperConfig viper 配置管理器
type ViperConfig struct {
	v *viper.Viper
	// 配置监听器
	listeners []ConfigChangeListener
	mu        sync.RWMutex
}

// ConfigChangeListener 配置变化监听器
type ConfigChangeListener func(oldValue, newValue interface{})

var (
	// 全局 viper 实例
	globalViperConfig *ViperConfig
	viperOnce         sync.Once
)

// GetViper 获取全局 viper 配置实例（单例）
func GetViper() *ViperConfig {
	viperOnce.Do(func() {
		v := viper.New()
		v.SetConfigType("yaml") // 设置配置类型为 YAML
		globalViperConfig = &ViperConfig{
			v:         v,
			listeners: make([]ConfigChangeListener, 0),
		}
	})
	return globalViperConfig
}

// InitViperFromNacos 从 nacos 初始化 viper 配置
// dataID: nacos 配置的 data ID
// group: nacos 配置的 group
func InitViperFromNacos(dataID, group string) error {
	vc := GetViper()

	// 从 dubbo-go 配置中心获取配置
	dynamicConfig := conf.GetEnvInstance().GetDynamicConfiguration()
	if dynamicConfig == nil {
		return fmt.Errorf("dynamic configuration not initialized")
	}

	// 获取配置内容
	content, err := dynamicConfig.GetProperties(dataID, config_center.WithGroup(group))
	if err != nil {
		return fmt.Errorf("failed to get config from nacos: %w", err)
	}

	if content == "" {
		return fmt.Errorf("config content is empty")
	}

	// 将配置内容设置到 viper
	if err := vc.v.ReadConfig(strings.NewReader(content)); err != nil {
		return fmt.Errorf("failed to read config into viper: %w", err)
	}

	logger.Infof("Viper initialized from nacos successfully, dataID=%s, group=%s", dataID, group)

	// 启动配置监听
	go vc.watchConfigChanges(dataID, group)

	return nil
}

// watchConfigChanges 监听 nacos 配置变化
func (vc *ViperConfig) watchConfigChanges(dataID, group string) {
	dynamicConfig := conf.GetEnvInstance().GetDynamicConfiguration()
	if dynamicConfig == nil {
		logger.Errorf("Failed to watch config: dynamic configuration is nil")
		return
	}

	// 创建配置监听器
	listener := &viperConfigListener{
		vc:     vc,
		dataID: dataID,
		group:  group,
	}

	// 添加监听器
	dynamicConfig.AddListener(dataID, listener, config_center.WithGroup(group))

	logger.Infof("Started watching nacos config changes, dataID=%s, group=%s", dataID, group)
}

// viperConfigListener viper 配置监听器
type viperConfigListener struct {
	vc     *ViperConfig
	dataID string
	group  string
}

// Process 实现 ConfigurationListener 接口
func (l *viperConfigListener) Process(event *config_center.ConfigChangeEvent) {
	logger.Infof("Nacos config changed: dataID=%s, group=%s, value=%v", l.dataID, l.group, event.Value)

	// 类型断言，将 Value 转换为字符串
	valueStr, ok := event.Value.(string)
	if !ok {
		logger.Errorf("Failed to convert config value to string: %v", event.Value)
		return
	}

	// 保存旧配置的副本
	oldConfig := l.vc.v.AllSettings()

	// 创建新的 viper 实例来读取新配置
	newViper := viper.New()
	newViper.SetConfigType("yaml")

	if err := newViper.ReadConfig(strings.NewReader(valueStr)); err != nil {
		logger.Errorf("Failed to read new config: %v", err)
		return
	}

	// 更新配置
	l.vc.mu.Lock()
	// 替换整个配置
	l.vc.v = newViper
	l.vc.mu.Unlock()

	// 触发配置变化回调
	newConfig := newViper.AllSettings()
	l.vc.notifyListeners(oldConfig, newConfig)

	logger.Infof("Viper config updated successfully")
}

// notifyListeners 通知所有监听器配置已变化
func (vc *ViperConfig) notifyListeners(oldConfig, newConfig map[string]interface{}) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	for _, listener := range vc.listeners {
		go listener(oldConfig, newConfig)
	}
}

// AddChangeListener 添加配置变化监听器
func (vc *ViperConfig) AddChangeListener(listener ConfigChangeListener) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.listeners = append(vc.listeners, listener)
}

// Get 获取配置值（支持点号路径，如 "redis.host"）
func (vc *ViperConfig) Get(key string) interface{} {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.Get(key)
}

// GetString 获取字符串配置
func (vc *ViperConfig) GetString(key string) string {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.GetString(key)
}

// GetInt 获取整数配置
func (vc *ViperConfig) GetInt(key string) int {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.GetInt(key)
}

// GetInt64 获取 int64 配置
func (vc *ViperConfig) GetInt64(key string) int64 {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.GetInt64(key)
}

// GetFloat64 获取 float64 配置
func (vc *ViperConfig) GetFloat64(key string) float64 {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.GetFloat64(key)
}

// GetBool 获取布尔配置
func (vc *ViperConfig) GetBool(key string) bool {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.GetBool(key)
}

// GetDuration 获取时间周期配置
func (vc *ViperConfig) GetDuration(key string) int {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	// viper 的 GetDuration 返回 time.Duration，这里简化为返回毫秒数
	return vc.v.GetInt(key)
}

// GetStringSlice 获取字符串数组配置
func (vc *ViperConfig) GetStringSlice(key string) []string {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.GetStringSlice(key)
}

// GetStringMap 获取 map 配置
func (vc *ViperConfig) GetStringMap(key string) map[string]interface{} {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.GetStringMap(key)
}

// GetStringMapString 获取 string map 配置
func (vc *ViperConfig) GetStringMapString(key string) map[string]string {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.GetStringMapString(key)
}

// IsSet 检查配置是否已设置
func (vc *ViperConfig) IsSet(key string) bool {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.IsSet(key)
}

// AllSettings 获取所有配置
func (vc *ViperConfig) AllSettings() map[string]interface{} {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.AllSettings()
}

// Unmarshal 解析配置到结构体
func (vc *ViperConfig) Unmarshal(rawVal interface{}) error {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.Unmarshal(rawVal)
}

// UnmarshalKey 解析指定 key 的配置到结构体
func (vc *ViperConfig) UnmarshalKey(key string, rawVal interface{}) error {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v.UnmarshalKey(key, rawVal)
}

// GetViper 获取底层 viper 实例（高级用法）
func (vc *ViperConfig) GetViper() *viper.Viper {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.v
}
