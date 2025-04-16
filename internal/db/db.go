package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
)

// GetOrCreateSchema 获取或创建 schema（数据库），返回固定加密的 schema_id
func GetOrCreateSchema(schemaName string, log *logrus.Logger) (string, error) {
	// 生成固定的 schema_id（基于 SHA-256 哈希）
	schemaID := generateSchemaID(schemaName)

	// 1. 检查 ClickHouse 中是否存在该数据库
	exists, err := DatabaseExists(schemaName, log)
	if err != nil {
		log.Error(fmt.Sprintf("检查 ClickHouse 数据库 %s 失败: %v", schemaName, err))
		return "", err
	}
	if !exists {
		// 如果数据库不存在，创建它
		if err := CreateDatabase(schemaName, log); err != nil {
			log.Error(fmt.Sprintf("创建 ClickHouse 数据库 %s 失败: %v", schemaName, err))
			return "", err
		}
	}

	// 2. 检查 BoltDB 是否有缓存
	storedID, err := GetSchemaIDByName(schemaName, log)
	if err != nil {
		log.Error(fmt.Sprintf("从 BoltDB 获取 schema %s 的 ID 失败: %v", schemaName, err))
		return "", err
	}

	// 3. 无论是否有缓存，都覆盖存储，确保一致性
	if storedID == "" {
		log.Info(fmt.Sprintf("BoltDB 中无 schema %s 的缓存，存储 ID: %s", schemaName, schemaID))
	}
	// 直接覆盖缓存，避免不一致
	if err := CacheSchema(schemaID, schemaName, log); err != nil {
		log.Error(fmt.Sprintf("缓存 schema %s 失败: %v", schemaName, err))
		// 缓存失败不影响返回，继续返回 schemaID
	}

	return schemaID, nil
}

// generateSchemaID 根据 schemaName 生成固定的加密 ID
func generateSchemaID(schemaName string) string {
	hash := sha256.Sum256([]byte(schemaName))
	return hex.EncodeToString(hash[:])
}
