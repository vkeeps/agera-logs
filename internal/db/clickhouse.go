package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/vkeeps/agera-logs/internal/model"
)

var (
	ClickHouseDB *sql.DB
	tablesMu     sync.Mutex
	tables       = make(map[string]bool)
)

const TablePrefix = "log_"

// InitClickHouse 初始化 ClickHouse 连接
func InitClickHouse() {
	addr := os.Getenv("CLICKHOUSE_ADDR")
	if addr == "" {
		addr = "localhost:29000"
	}
	user := os.Getenv("CLICKHOUSE_USER")
	if user == "" {
		user = "default"
	}
	pass := os.Getenv("CLICKHOUSE_PASS")
	if pass == "" {
		pass = "crane"
	}
	dbName := os.Getenv("CLICKHOUSE_DB")
	if dbName == "" {
		dbName = "default"
	}

	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: dbName,
			Username: user,
			Password: pass,
		},
	})
	ClickHouseDB = conn
	if err := conn.Ping(); err != nil {
		log.Fatalf("ClickHouse 连不上: %v", err)
	}
	log.Printf("ClickHouse 连接成功，地址: %s", addr)
}

// EnsureTable 在指定 schema（数据库）中创建表，添加字段约束
func EnsureTable(schemaName, moduleName string) error {
	tablesMu.Lock()
	defer tablesMu.Unlock()

	tableName := fmt.Sprintf("%s.%s%s_%s", schemaName, TablePrefix, schemaName, moduleName)
	if tables[tableName] {
		return nil
	}

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			output String NOT NULL,
			detail String,
			error_info String,
			service String NOT NULL,
			client_ip String NOT NULL,
			client_addr String NOT NULL,
			operator String NOT NULL,
			operation_time DateTime NOT NULL
		) ENGINE = MergeTree()
		ORDER BY (operation_time)
	`, tableName)
	_, err := ClickHouseDB.Exec(query)
	if err != nil {
		return fmt.Errorf("创建表 %s 失败: %v", tableName, err)
	}
	tables[tableName] = true
	log.Printf("表 %s 创建成功", tableName)
	return nil
}

// InsertLog 插入日志到指定 schema 的表
func InsertLog(entry *model.Log) error {
	if err := EnsureTable(string(entry.Schema), string(entry.Module)); err != nil {
		return err
	}

	// 确保 NOT NULL 字段有默认值
	logEntry := model.LogEntry{
		LogBase: model.LogBase{
			Output:     entry.Output,
			Detail:     entry.Detail,
			ErrorInfo:  entry.ErrorInfo,
			Service:    nonEmpty(entry.Service, "unknown"),
			ClientIP:   nonEmpty(entry.ClientIP, "0.0.0.0"),
			ClientAddr: nonEmpty(entry.ClientAddr, "unknown"),
		},
		Operator:      nonEmpty("unknown", "unknown"), // 默认值，需从推送数据获取
		OperationTime: entry.Timestamp,
	}

	tableName := fmt.Sprintf("%s.%s%s_%s", entry.Schema, TablePrefix, entry.Schema, entry.Module)
	query := fmt.Sprintf(`
		INSERT INTO %s (output, detail, error_info, service, client_ip, client_addr, operator, operation_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName)
	_, err := ClickHouseDB.Exec(query,
		logEntry.Output, logEntry.Detail, logEntry.ErrorInfo, logEntry.Service,
		logEntry.ClientIP, logEntry.ClientAddr, logEntry.Operator, logEntry.OperationTime)
	if err != nil {
		log.Printf("插入日志到 %s 失败: %v", tableName, err)
		return err
	}
	return nil
}

// nonEmpty 返回非空值，若输入为空则使用默认值
func nonEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// DatabaseExists 检查 ClickHouse 中是否存在指定数据库
func DatabaseExists(dbName string) (bool, error) {
	query := fmt.Sprintf("EXISTS DATABASE %s", dbName)
	var exists uint8
	err := ClickHouseDB.QueryRow(query).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("检查数据库 %s 存在性失败: %v", dbName, err)
	}
	return exists == 1, nil
}

// CreateDatabase 创建 ClickHouse 数据库
func CreateDatabase(dbName string) error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)
	_, err := ClickHouseDB.Exec(query)
	if err != nil {
		return fmt.Errorf("创建数据库 %s 失败: %v", dbName, err)
	}
	log.Printf("数据库 %s 创建成功", dbName)
	return nil
}
