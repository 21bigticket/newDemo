package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dubbogo/gost/log/logger"
)

// NacosConfig Nacos 配置结构体
type NacosConfig struct {
	Address   string // Nacos 服务器地址
	Namespace string // 命名空间
	Group     string // 分组
	DataID    string // 配置 Data ID
	Timeout   string // 超时时间
}

// Config 应用配置结构体
type Config struct {
	Nacos    NacosConfig
	AppName  string
	AppPort  int
	LogLevel string
}

// defaultNacosConfig 默认 Nacos 配置
var defaultNacosConfig = NacosConfig{
	Address:   "127.0.0.1:8848",
	Namespace: "public",
	Group:     "DEFAULT_GROUP",
	DataID:    "",
	Timeout:   "3s",
}

// ParseConfig 解析配置，优先级：命令行参数 > 环境变量 > 默认值
func ParseConfig(appType string) (*Config, error) {
	// 定义命令行参数
	var (
		nacosAddr   = flag.String("nacos-addr", "", "Nacos server address")
		namespace   = flag.String("namespace", "", "Nacos namespace")
		group       = flag.String("group", "", "Nacos group")
		dataID      = flag.String("data-id", "", "Nacos config data ID")
		timeout     = flag.String("timeout", "", "Nacos timeout")
		appName     = flag.String("app-name", "", "Application name")
		appPort     = flag.Int("port", 0, "Application port")
		logLevel    = flag.String("log-level", "", "Log level")
		showVersion = flag.Bool("version", false, "Show version")
		help        = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	// 处理帮助和版本信息
	if *help {
		printHelp(appType)
		//os.Exit(0)
	}
	if *showVersion {
		fmt.Printf("%s version 1.0.0\n", appType)
		os.Exit(0)
	}

	// 构建配置，按照优先级处理
	config := &Config{
		Nacos: defaultNacosConfig,
	}

	// 设置应用相关配置
	config.AppName = getStringValue(*appName, getEnv("APP_NAME"), "")
	config.AppPort = getIntValue(*appPort, getEnvInt("APP_PORT"), 20001)
	config.LogLevel = getStringValue(*logLevel, getEnv("LOG_LEVEL"), "info")

	// 设置 Nacos 相关配置
	config.Nacos.Address = getStringValue(*nacosAddr, getEnv("NACOS_ADDR"), defaultNacosConfig.Address)
	config.Nacos.Namespace = getStringValue(*namespace, getEnv("NACOS_NAMESPACE"), defaultNacosConfig.Namespace)
	config.Nacos.Group = getStringValue(*group, getEnv("NACOS_GROUP"), defaultNacosConfig.Group)
	config.Nacos.Timeout = getStringValue(*timeout, getEnv("NACOS_TIMEOUT"), defaultNacosConfig.Timeout)

	// DataID 需要特殊处理
	if *dataID != "" {
		config.Nacos.DataID = *dataID
	} else if envDataID := getEnv("NACOS_DATA_ID"); envDataID != "" {
		config.Nacos.DataID = envDataID
	} else {
		config.Nacos.DataID = fmt.Sprintf("%s-config", strings.ToLower(config.AppName))
	}

	// 验证必要配置
	if config.Nacos.Address == "" {
		return nil, fmt.Errorf("nacos address is required")
	}

	return config, nil
}

// getStringValue 获取字符串值，按优先级：命令行 > 环境变量 > 默认值
func getStringValue(flagVal, envVal, defaultVal string) string {
	if flagVal != "" {
		return flagVal
	}
	if envVal != "" {
		return envVal
	}
	return defaultVal
}

// getIntValue 获取整数值，按优先级：命令行 > 环境变量 > 默认值
func getIntValue(flagVal, envVal, defaultVal int) int {
	if flagVal != 0 {
		return flagVal
	}
	if envVal != 0 {
		return envVal
	}
	return defaultVal
}

// getEnv 获取环境变量值
func getEnv(key string) string {
	return os.Getenv(key)
}

// getEnvInt 获取环境变量整数值
func getEnvInt(key string) int {
	val := os.Getenv(key)
	if val == "" {
		return 0
	}

	// 简单的字符串转整数，实际项目中可以使用 strconv
	result := 0
	for _, c := range val {
		if c < '0' || c > '9' {
			return 0
		}
		result = result*10 + int(c-'0')
	}
	return result
}

// printHelp 使用logger打印帮助信息
func printHelp(appType string) {
	logger.Info(fmt.Sprintf("Usage: %s [options]", appType))
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  -nacos-addr string    Nacos server address (e.g., 192.168.139.230:8848)")
	logger.Info("  -namespace string     Nacos namespace (default: public)")
	logger.Info("  -group string         Nacos group (default: DEFAULT_GROUP)")
	logger.Info("  -data-id string       Nacos config data ID")
	logger.Info("  -timeout string       Nacos timeout (default: 3s)")
	logger.Info("  -app-name string      Application name")
	logger.Info(fmt.Sprintf("  -port int             Application port (server default: 20001)"))
	logger.Info("  -log-level string     Log level (default: info)")
	logger.Info("  -version              Show version")
	logger.Info("  -help                 Show this help")
	logger.Info("")
	logger.Info("Environment Variables:")
	logger.Info("  NACOS_ADDR            Nacos server address")
	logger.Info("  NACOS_NAMESPACE       Nacos namespace")
	logger.Info("  NACOS_GROUP           Nacos group")
	logger.Info("  NACOS_DATA_ID         Nacos config data ID")
	logger.Info("  NACOS_TIMEOUT         Nacos timeout")
	logger.Info("  APP_NAME              Application name")
	logger.Info("  APP_PORT              Application port")
	logger.Info("  LOG_LEVEL             Log level")
	logger.Info("")
	logger.Info("Priority: Command Line > Environment Variables > Defaults")
}
