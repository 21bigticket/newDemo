package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/dubbogo/gost/log/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// MySQLConfig MySQL 配置结构体
type MySQLConfig struct {
	Host            string        `json:"host" yaml:"host"`
	Port            int           `json:"port" yaml:"port"`
	Username        string        `json:"username" yaml:"username"`
	Password        string        `json:"password" yaml:"password"`
	Database        string        `json:"database" yaml:"database"`
	Charset         string        `json:"charset" yaml:"charset"`
	Location        string        `json:"location" yaml:"location"`
	ConnTimeout     time.Duration `json:"conn_timeout" yaml:"conn_timeout"`
	ReadTimeout     time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout" yaml:"write_timeout"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
}

// CreateDB 创建 GORM 数据库连接
func (mc *MySQLConfig) CreateDB() (*gorm.DB, error) {
	dsn := mc.DSN()

	logger.Infof("MySQL DSN: %s:***@tcp(%s:%d)/%s", mc.Username, mc.Host, mc.Port, mc.Database)

	// 配置 GORM
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 禁用外键约束
		DisableForeignKeyConstraintWhenMigrating: true,
		// 跳过默认事务
		SkipDefaultTransaction: true,
		// 预编译SQL
		PrepareStmt: true,
		// 日志配置
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池参数
	if mc.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(mc.MaxIdleConns)
	}
	if mc.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(mc.MaxOpenConns)
	}

	// 设置连接最大存活时间
	if mc.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(mc.ConnMaxLifetime)
	}

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Infof("MySQL connected successfully: %s@%s:%d/%s",
		mc.Username, mc.Host, mc.Port, mc.Database)

	return db, nil
}

// DSN 生成MySQL连接字符串（带参数转义）
func (mc *MySQLConfig) DSN() string {
	// 转义特殊字符
	username := url.QueryEscape(mc.Username)
	password := url.QueryEscape(mc.Password)
	location := url.QueryEscape(mc.Location)

	// 构建基础DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
		username, password, mc.Host, mc.Port, mc.Database, mc.Charset, location)

	// 添加超时参数
	timeoutParams := []string{}

	// 连接超时（默认10秒）
	if mc.ConnTimeout > 0 {
		timeoutParams = append(timeoutParams, fmt.Sprintf("timeout=%ds", int(mc.ConnTimeout.Seconds())))
	} else {
		timeoutParams = append(timeoutParams, "timeout=10s")
	}

	// 读取超时（默认30秒）
	if mc.ReadTimeout > 0 {
		timeoutParams = append(timeoutParams, fmt.Sprintf("readTimeout=%ds", int(mc.ReadTimeout.Seconds())))
	} else {
		timeoutParams = append(timeoutParams, "readTimeout=30s")
	}

	// 写入超时（默认30秒）
	if mc.WriteTimeout > 0 {
		timeoutParams = append(timeoutParams, fmt.Sprintf("writeTimeout=%ds", int(mc.WriteTimeout.Seconds())))
	} else {
		timeoutParams = append(timeoutParams, "writeTimeout=30s")
	}

	// 添加超时参数到DSN
	if len(timeoutParams) > 0 {
		dsn += "&" + timeoutParams[0]
		for i := 1; i < len(timeoutParams); i++ {
			dsn += "&" + timeoutParams[i]
		}
	}

	return dsn
}

// parseDurationValue 解析duration值（支持多种类型）
func parseDurationValue(value interface{}) time.Duration {
	switch v := value.(type) {
	case string:
		return parseDuration(v)
	case int:
		return time.Duration(v) * time.Second
	case int64:
		return time.Duration(v) * time.Second
	case float64:
		return time.Duration(v) * time.Second
	case float32:
		return time.Duration(v) * time.Second
	default:
		logger.Errorf("Unsupported duration type %T for value %v", value, value)
		return 0
	}
}

// GetMySQLConfigFromDubbo 从 dubbo-go 配置中心获取 MySQL 配置
func GetMySQLConfigFromDubbo() (*MySQLConfig, error) {
	// 从配置管理器获取配置
	mysqlMap := GetStringMap("mysql")
	if mysqlMap == nil {
		return nil, fmt.Errorf("mysql config not found")
	}

	config := &MySQLConfig{}

	// 解析各个字段
	if v, ok := mysqlMap["host"].(string); ok {
		config.Host = v
	}
	if v, ok := mysqlMap["port"].(int); ok {
		config.Port = v
	}
	if v, ok := mysqlMap["username"].(string); ok {
		config.Username = v
	}
	if v, ok := mysqlMap["password"].(string); ok {
		config.Password = v
	}
	if v, ok := mysqlMap["database"].(string); ok {
		config.Database = v
	}
	if v, ok := mysqlMap["charset"].(string); ok {
		config.Charset = v
	}
	if v, ok := mysqlMap["location"].(string); ok {
		config.Location = v
	}

	// 解析超时配置
	if v, ok := mysqlMap["conn_timeout"]; ok {
		config.ConnTimeout = parseDurationValue(v)
	}
	if v, ok := mysqlMap["read_timeout"]; ok {
		config.ReadTimeout = parseDurationValue(v)
	}
	if v, ok := mysqlMap["write_timeout"]; ok {
		config.WriteTimeout = parseDurationValue(v)
	}
	if v, ok := mysqlMap["conn_max_lifetime"]; ok {
		config.ConnMaxLifetime = parseDurationValue(v)
	}

	// 解析连接池配置
	if v, ok := mysqlMap["max_idle_conns"].(int); ok {
		config.MaxIdleConns = v
	}
	if v, ok := mysqlMap["max_open_conns"].(int); ok {
		config.MaxOpenConns = v
	}

	// 设置默认值
	if config.Charset == "" {
		config.Charset = "utf8mb4"
	}
	if config.Location == "" {
		config.Location = "Local"
	}
	if config.Port == 0 {
		config.Port = 3306
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 10
	}
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 100
	}

	logger.Infof("Loaded MySQL config: %+v", config)
	return config, nil
}
