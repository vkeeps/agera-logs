package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vkeeps/agera-logs/internal/db"
	"github.com/vkeeps/agera-logs/internal/model"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/logs", createLog)
	r.GET("/logs/:schema", getLogs)
	r.POST("/schemas", createSchema)
	r.GET("/schemas/:name", getSchema)

	return r
}

// createSchema 创建 schema 并返回 schema_id
func createSchema(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据格式有误"})
		return
	}

	// 生成唯一的 schema_id
	schemaID := uuid.New().String()

	// 确保 ClickHouse 中存在对应的表
	if err := db.EnsureTable(req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建表失败"})
		return
	}

	// 将 schema_id 和 schema 名称的映射存入 BoltDB
	if err := db.CacheSchema(schemaID, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "缓存 schema 失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schema": req.Name, "id": schemaID})
}

// getSchema 根据 schema 名称获取 schema_id
func getSchema(c *gin.Context) {
	schemaName := c.Param("name")
	schemaID, err := db.GetSchemaIDByName(schemaName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询 schema 失败"})
		return
	}
	if schemaID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema 未找到"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema": schemaName, "id": schemaID})
}

// createLog 创建日志（HTTP 方式）
func createLog(c *gin.Context) {
	var entry model.Log
	if err := c.BindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据格式有误"})
		return
	}

	// 设置时间戳和 PushType
	entry.Timestamp = time.Now()
	entry.PushType = model.PushTypeHTTP

	if err := db.InsertLog(&entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "日志插入失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "日志已添加"})
}

// getLogs 获取指定 schema 的日志
func getLogs(c *gin.Context) {
	schema := c.Param("schema")
	tableName := db.TablePrefix + schema

	// 确保表存在
	if err := db.EnsureTable(schema); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "表不存在或创建失败"})
		return
	}

	// 查询日志
	rows, err := db.ClickHouseDB.Query("SELECT * FROM " + tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询日志失败"})
		return
	}
	defer rows.Close()

	var logs []model.Log
	for rows.Next() {
		var entry model.Log
		if err := rows.Scan(&entry.Timestamp, &entry.Module, &entry.Output, &entry.Detail,
			&entry.ErrorInfo, &entry.Service, &entry.ClientIP, &entry.ClientAddr, &entry.PushType); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "日志解析失败"})
			return
		}
		logs = append(logs, entry)
	}
	c.JSON(http.StatusOK, logs)
}
