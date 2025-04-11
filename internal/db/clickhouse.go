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
	tables       = make(map[string]bool) // 缓存已创建的表
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

// EnsureTable 确保表存在
func EnsureTable(schemaName string) error {
	tablesMu.Lock()
	defer tablesMu.Unlock()

	tableName := TablePrefix + schemaName
	if tables[tableName] {
		return nil
	}

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			timestamp DateTime,
			module String,
			output String,
			detail String,
			error_info String,
			service String,
			client_ip String,
			client_addr String,
			push_type String
		) ENGINE = MergeTree()
		ORDER BY (timestamp)
	`, tableName)
	_, err := ClickHouseDB.Exec(query)
	if err != nil {
		return fmt.Errorf("创建表 %s 失败: %v", tableName, err)
	}
	tables[tableName] = true
	log.Printf("表 %s 创建成功", tableName)
	return nil
}

// InsertLog 插入日志
func InsertLog(entry *model.Log) error {
	if err := EnsureTable(string(entry.Schema)); err != nil {
		return err
	}

	tableName := TablePrefix + string(entry.Schema)
	query := fmt.Sprintf(`
		INSERT INTO %s (timestamp, module, output, detail, error_info, service, client_ip, client_addr, push_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName)
	_, err := ClickHouseDB.Exec(query,
		entry.Timestamp, entry.Module, entry.Output, entry.Detail, entry.ErrorInfo,
		entry.Service, entry.ClientIP, entry.ClientAddr, entry.PushType)
	return err
}

// SchemaExists 检查 ClickHouse 中是否存在指定 schema 表
func SchemaExists(schemaName string) (bool, error) {
	tableName := TablePrefix + schemaName
	query := fmt.Sprintf("EXISTS TABLE %s", tableName)
	var exists uint8
	err := ClickHouseDB.QueryRow(query).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("检查表 %s 存在性失败: %v", tableName, err)
	}
	return exists == 1, nil
}

// CreateSchema 在 ClickHouse 中创建 schema 表
func CreateSchema(schemaName string) error {
	return EnsureTable(schemaName) // 复用 EnsureTable 逻辑
}
