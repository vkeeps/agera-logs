package db

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sirupsen/logrus"
	"github.com/vkeeps/agera-logs/internal/model"
	"os"
)

var (
	ClickHouseDB *sql.DB
	tablesMu     sync.Mutex
	tables       = make(map[string]bool)
)

const TablePrefix = "log_"

// InitClickHouse 初始化 ClickHouse 连接
func InitClickHouse(log *logrus.Logger) {
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
		log.Fatal(fmt.Sprintf("ClickHouse 连不上: %v", err))
	}
	log.Info(fmt.Sprintf("ClickHouse 连接成功，地址: %s", addr))
}

// EnsureTable 在指定 schema（数据库）中创建表，添加字段约束
func EnsureTable(schemaName, moduleName string, log *logrus.Logger) error {
	tablesMu.Lock()
	defer tablesMu.Unlock()

	tableName := fmt.Sprintf("%s.%s%s", schemaName, TablePrefix, moduleName)
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
			operator_id String NOT NULL,
			operator String NOT NULL,
			operation_time DateTime NOT NULL,
			push_type String NOT NULL
		) ENGINE = MergeTree()
		ORDER BY (operation_time)
	`, tableName)
	_, err := ClickHouseDB.Exec(query)
	if err != nil {
		log.Error(fmt.Sprintf("创建表 %s 失败: %v", tableName, err))
		return fmt.Errorf("创建表 %s 失败: %v", tableName, err)
	}
	tables[tableName] = true
	log.Info(fmt.Sprintf("表 %s 创建成功", tableName))
	return nil
}

// InsertLogs 批量插入日志
func InsertLogs(entries []*model.Log, log *logrus.Logger) error {
	if len(entries) == 0 {
		return nil
	}

	// 确保表存在（使用第一个日志的 schema 和 module）
	schemaName := string(entries[0].Schema)
	moduleName := string(entries[0].Module)
	if err := EnsureTable(schemaName, moduleName, log); err != nil {
		return err
	}

	// 开始事务
	tx, err := ClickHouseDB.Begin()
	if err != nil {
		log.Error(fmt.Sprintf("开始 ClickHouse 事务失败: %v", err))
		return fmt.Errorf("开始 ClickHouse 事务失败: %v", err)
	}

	// 准备批量插入语句
	tableName := fmt.Sprintf("%s.%s%s", schemaName, TablePrefix, moduleName)
	query := fmt.Sprintf(`
		INSERT INTO %s (output, detail, error_info, service, client_ip, client_addr, operator_id, operator, operation_time, push_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName)
	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		log.Error(fmt.Sprintf("准备 ClickHouse 插入语句失败: %v", err))
		return fmt.Errorf("准备 ClickHouse 插入语句失败: %v", err)
	}

	// 批量插入数据
	for _, entry := range entries {
		_, err := stmt.Exec(
			entry.Output,
			entry.Detail,
			entry.ErrorInfo,
			nonEmpty(entry.Service, "unknown"),
			nonEmpty(entry.ClientIP, "0.0.0.0"),
			nonEmpty(entry.ClientAddr, "unknown"),
			nonEmpty(entry.OperatorID, "unknown"),
			nonEmpty(entry.Operator, "unknown"),
			entry.Timestamp,
			string(entry.PushType),
		)
		if err != nil {
			tx.Rollback()
			log.Error(fmt.Sprintf("执行 ClickHouse 插入失败: %v", err))
			return fmt.Errorf("执行 ClickHouse 插入失败: %v", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Error(fmt.Sprintf("提交 ClickHouse 事务失败: %v", err))
		return fmt.Errorf("提交 ClickHouse 事务失败: %v", err)
	}
	return nil
}

// InsertLog 单条插入日志，调用 InsertLogs
func InsertLog(entry *model.Log, log *logrus.Logger) error {
	return InsertLogs([]*model.Log{entry}, log)
}

// nonEmpty 返回非空值，若输入为空则使用默认值
func nonEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// DatabaseExists 检查 ClickHouse 中是否存在指定数据库
func DatabaseExists(dbName string, log *logrus.Logger) (bool, error) {
	query := fmt.Sprintf("EXISTS DATABASE %s", dbName)
	var exists uint8
	err := ClickHouseDB.QueryRow(query).Scan(&exists)
	if err != nil {
		log.Error(fmt.Sprintf("检查数据库 %s 存在性失败: %v", dbName, err))
		return false, fmt.Errorf("检查数据库 %s 存在性失败: %v", dbName, err)
	}
	return exists == 1, nil
}

// CreateDatabase 创建 ClickHouse 数据库
func CreateDatabase(dbName string, log *logrus.Logger) error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)
	_, err := ClickHouseDB.Exec(query)
	if err != nil {
		log.Error(fmt.Sprintf("创建数据库 %s 失败: %v", dbName, err))
		return fmt.Errorf("创建数据库 %s 失败: %v", dbName, err)
	}
	log.Info(fmt.Sprintf("数据库 %s 创建成功", dbName))
	return nil
}
