package db

import (
	"log"
)

// GetOrCreateSchema 获取或创建 schema，返回唯一标识
func GetOrCreateSchema(schemaName string) (string, error) {
	// 1. 从 BoltDB 获取 schemaID
	schemaID, err := GetSchemaIDByName(schemaName) // 修正为 GetSchemaIDByName
	if err != nil {
		return "", err
	}
	if schemaID != "" {
		return schemaID, nil
	}

	// 2. 检查 ClickHouse 中是否存在该 schema
	exists, err := SchemaExists(schemaName)
	if err != nil {
		return "", err
	}
	if exists {
		// 如果存在，直接使用 schemaName 作为 schemaID 并缓存
		schemaID = schemaName
		err = CacheSchema(schemaName, schemaID)
		if err != nil {
			log.Printf("缓存 schema %s 失败: %v", schemaName, err)
		}
		return schemaID, nil
	}

	// 3. 创建 schema
	err = CreateSchema(schemaName)
	if err != nil {
		return "", err
	}

	// 4. 使用 schemaName 作为 schemaID 并缓存
	schemaID = schemaName
	err = CacheSchema(schemaName, schemaID)
	if err != nil {
		log.Printf("缓存 schema %s 失败: %v", schemaName, err)
	}

	return schemaID, nil
}
