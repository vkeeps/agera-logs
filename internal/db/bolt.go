package db

import (
	"log"

	"github.com/boltdb/bolt"
)

var BoltDB *bolt.DB

// InitBolt 初始化 BoltDB
func InitBolt() {
	var err error
	BoltDB, err = bolt.Open("logsvc_config.db", 0600, nil)
	if err != nil {
		log.Fatalf("BoltDB 打不开: %v", err)
	}
	err = BoltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("schemas"))
		return err
	})
	if err != nil {
		log.Fatalf("schemas 桶创建失败: %v", err)
	}
	log.Println("BoltDB 初始化成功")
}

// CacheSchema 将 schema_id 和 schema 名称的映射存入 BoltDB
func CacheSchema(schemaID, schemaName string) error {
	return BoltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("schemas"))
		return b.Put([]byte(schemaID), []byte(schemaName))
	})
}

// GetSchemaNameByID 根据 schema_id 获取 schema 名称
func GetSchemaNameByID(schemaID string) (string, error) {
	var schemaName string
	err := BoltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("schemas"))
		name := b.Get([]byte(schemaID))
		if name != nil {
			schemaName = string(name)
		}
		return nil
	})
	return schemaName, err
}

// GetSchemaIDByName 根据 schema 名称获取 schema_id
func GetSchemaIDByName(schemaName string) (string, error) {
	var schemaID string
	err := BoltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("schemas"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if string(v) == schemaName {
				schemaID = string(k)
				break
			}
		}
		return nil
	})
	return schemaID, err
}
