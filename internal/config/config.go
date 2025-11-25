package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	RPC      RPCConfig
	Wallet   WalletConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string // memory, postgres, mysql
	DSN      string // 完整的 DSN 连接字符串（如果设置，优先使用）
	Host     string // 数据库主机地址
	Port     int    // 数据库端口
	User     string // 数据库用户名
	Password string // 数据库密码
	Database string // 数据库名称
	SSLMode  string // SSL 模式（PostgreSQL 使用，默认 disable）
}

// RPCConfig RPC 配置
type RPCConfig struct {
	Ethereum string
	Bitcoin  string
	Solana   string
}

// WalletConfig 钱包配置
type WalletConfig struct {
	MasterSeed    string // 主种子（实际应该从环境变量或密钥管理服务获取）
	MasterKeyPath string // 主密钥文件路径
}

// Load 加载配置
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8081),
		},
		Database: DatabaseConfig{
			Type:     getEnv("DB_TYPE", "postgres"),
			DSN:      getEnv("DB_DSN", ""),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "123456"),
			Database: getEnv("DB_NAME", "wallet"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		RPC: RPCConfig{
			Ethereum: getEnv("RPC_ETHEREUM", "https://eth.llamarpc.com"),
			Bitcoin:  getEnv("RPC_BITCOIN", ""),
			Solana:   getEnv("RPC_SOLANA", ""),
		},
		Wallet: WalletConfig{
			MasterSeed:    getEnv("WALLET_MASTER_SEED", ""),
			MasterKeyPath: getEnv("WALLET_MASTER_KEY_PATH", ""),
		},
	}
}

// Address 返回服务器地址
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// BuildDSN 构建数据库连接字符串
// 如果 DSN 已设置，直接返回；否则根据配置字段构建
func (c *DatabaseConfig) BuildDSN() string {
	// 如果已经设置了完整的 DSN，直接使用
	if c.DSN != "" {
		return c.DSN
	}

	// 根据数据库类型构建 DSN
	switch c.Type {
	case "postgres", "postgresql":
		// PostgreSQL DSN 格式: postgres://user:password@host:port/database?sslmode=disable
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
	case "mysql":
		// MySQL DSN 格式: user:password@tcp(host:port)/database?parseTime=true
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
			c.User, c.Password, c.Host, c.Port, c.Database)
	default:
		// memory 或其他类型，返回空字符串
		return ""
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

