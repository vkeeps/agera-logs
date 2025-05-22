package db

import (
	"database/sql"
	"fmt"
	"strings"
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
			log_level String NOT NULL,
			operator_id String NOT NULL,
			operator String NOT NULL,
			operator_ip String NOT NULL,
			operator_equipment String NOT NULL,
			operator_company String NOT NULL,
			operator_project String NOT NULL,
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
	tableName := fmt.Sprintf("%s.%s%s_%s", schemaName, TablePrefix, schemaName, moduleName)
	query := fmt.Sprintf(`
		INSERT INTO %s (output, detail, error_info, service, client_ip, client_addr, log_level, operator_id, operator, operator_ip, operator_equipment, operator_company, operator_project, operation_time, push_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName)
	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		log.Error(fmt.Sprintf("准备 ClickHouse 插入语句失败: %v", err))
		return fmt.Errorf("准备 ClickHouse 插入语句失败: %v", err)
	}

	// 批量插入数据
	for _, entry := range entries {
		// LogLevel特殊处理：为空时设为INFO，并确保大写
		logLevel := entry.LogLevel
		if logLevel == "" {
			logLevel = "INFO"
		} else {
			logLevel = strings.ToUpper(logLevel)
		}

		_, err := stmt.Exec(
			entry.Output,
			entry.Detail,
			entry.ErrorInfo,
			nonEmpty(entry.Service, "unknown"),
			nonEmpty(entry.ClientIP, "0.0.0.0"),
			nonEmpty(entry.ClientAddr, "unknown"),
			logLevel, // 使用处理后的logLevel
			nonEmpty(entry.OperatorID, "unknown"),
			nonEmpty(entry.Operator, "unknown"),
			nonEmpty(entry.OperatorIP, "unknown"),
			nonEmpty(entry.OperatorEquipment, "unknown"),
			nonEmpty(entry.OperatorCompany, "unknown"),
			nonEmpty(entry.OperatorProject, "unknown"),
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

// GetAllSchemas 获取所有 schema（数据库）列表
func GetAllSchemas(log *logrus.Logger) ([]map[string]interface{}, error) {
	rows, err := ClickHouseDB.Query("SHOW DATABASES")
	if err != nil {
		log.Error(fmt.Sprintf("查询所有数据库失败: %v", err))
		return nil, fmt.Errorf("查询所有数据库失败: %v", err)
	}
	defer rows.Close()

	var schemas []map[string]interface{}
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			log.Error(fmt.Sprintf("解析数据库名称失败: %v", err))
			return nil, fmt.Errorf("解析数据库名称失败: %v", err)
		}
		if dbName == "system" || dbName == "default" {
			continue // 跳过系统数据库
		}
		schemaID, err := GetOrCreateSchema(dbName, log)
		if err != nil {
			log.Error(fmt.Sprintf("获取 schema %s 的 ID 失败: %v", dbName, err))
			continue
		}
		schemas = append(schemas, map[string]interface{}{
			"name": dbName,
			"id":   schemaID,
		})
	}
	return schemas, nil
}

// GetTablesBySchemaId 根据 schemaId 获取所有相关表
func GetTablesBySchemaId(schemaId string, log *logrus.Logger) ([]string, error) {
	// 先通过schemaId获取实际的数据库名称
	schemaName, err := GetSchemaNameByID(schemaId, log)
	if err != nil {
		log.Error(fmt.Sprintf("获取 schema_id %s 对应的数据库名失败: %v", schemaId, err))
		return nil, fmt.Errorf("获取 schema_id %s 对应的数据库名失败: %v", schemaId, err)
	}

	if schemaName == "" {
		log.Error(fmt.Sprintf("未找到 schema_id %s 对应的数据库名", schemaId))
		return nil, fmt.Errorf("未找到 schema_id %s 对应的数据库名", schemaId)
	}

	rows, err := ClickHouseDB.Query("SELECT name FROM system.tables WHERE database = ?", schemaName)
	if err != nil {
		log.Error(fmt.Sprintf("查询 schema %s 的表失败: %v", schemaName, err))
		return nil, fmt.Errorf("查询 schema %s 的表失败: %v", schemaName, err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Error(fmt.Sprintf("解析表名称失败: %v", err))
			return nil, fmt.Errorf("解析表名称失败: %v", err)
		}
		if strings.HasPrefix(tableName, TablePrefix) {
			tables = append(tables, fmt.Sprintf("%s.%s", schemaName, tableName))
		}
	}
	return tables, nil
}

// GetModulesBySchemaId 根据 schemaId 获取所有相关模块
func GetModulesBySchemaId(schemaId string, log *logrus.Logger) ([]string, error) {
	// 先通过schemaId获取实际的数据库名称
	schemaName, err := GetSchemaNameByID(schemaId, log)
	if err != nil {
		log.Error(fmt.Sprintf("获取 schema_id %s 对应的数据库名失败: %v", schemaId, err))
		return nil, fmt.Errorf("获取 schema_id %s 对应的数据库名失败: %v", schemaId, err)
	}

	if schemaName == "" {
		log.Error(fmt.Sprintf("未找到 schema_id %s 对应的数据库名", schemaId))
		return nil, fmt.Errorf("未找到 schema_id %s 对应的数据库名", schemaId)
	}

	// 直接查询数据库表，而不是通过GetTablesBySchemaId
	rows, err := ClickHouseDB.Query("SELECT name FROM system.tables WHERE database = ?", schemaName)
	if err != nil {
		log.Error(fmt.Sprintf("查询 schema %s 的表失败: %v", schemaName, err))
		return nil, fmt.Errorf("查询 schema %s 的表失败: %v", schemaName, err)
	}
	defer rows.Close()

	var modules []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Error(fmt.Sprintf("解析表名称失败: %v", err))
			return nil, fmt.Errorf("解析表名称失败: %v", err)
		}

		prefix := TablePrefix + schemaName + "_"
		if strings.HasPrefix(tableName, prefix) {
			module := strings.TrimPrefix(tableName, prefix)
			modules = append(modules, module)
		}
	}
	return modules, nil
}
