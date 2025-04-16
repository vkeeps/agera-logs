package db

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
)

var BoltDB *bolt.DB

// InitBolt 初始化 BoltDB
func InitBolt(log *logrus.Logger) {
	var err error
	BoltDB, err = bolt.Open("logsvc_config.db", 0600, nil)
	if err != nil {
		log.Fatal(fmt.Sprintf("BoltDB 打不开: %v", err))
	}
	err = BoltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("schemas"))
		return err
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("schemas 桶创建失败: %v", err))
	}
	log.Info("BoltDB 初始化成功")
}

// CacheSchema 将 schema_name 和 schema_id 的映射存入 BoltDB
func CacheSchema(schemaID, schemaName string, log *logrus.Logger) error {
	return BoltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("schemas"))
		return b.Put([]byte(schemaName), []byte(schemaID))
	})
}

// GetSchemaNameByID 根据 schema_id 获取 schema 名称
func GetSchemaNameByID(schemaID string, log *logrus.Logger) (string, error) {
	var schemaName string
	err := BoltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("schemas"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if string(v) == schemaID {
				schemaName = string(k)
				break
			}
		}
		return nil
	})
	return schemaName, err
}

// GetSchemaIDByName 根据 schema 名称获取 schema_id
func GetSchemaIDByName(schemaName string, log *logrus.Logger) (string, error) {
	var schemaID string
	err := BoltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("schemas"))
		id := b.Get([]byte(schemaName))
		if id != nil {
			schemaID = string(id)
		}
		return nil
	})
	return schemaID, err
}
