package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vkeeps/agera-logs/internal/db"
	"github.com/vkeeps/agera-logs/internal/model"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/logs", createLog)
	r.GET("/logs/:schema/:module", getLogs)
	r.POST("/schemas", createSchema)
	r.GET("/schemas/:name", getSchema)

	return r
}

func createSchema(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据格式有误"})
		return
	}

	schemaID, err := db.GetOrCreateSchema(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建 schema 失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schema": req.Name, "id": schemaID})
}

func getSchema(c *gin.Context) {
	schemaName := c.Param("name")
	schemaID, err := db.GetOrCreateSchema(schemaName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询 schema 失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema": schemaName, "id": schemaID})
}

func createLog(c *gin.Context) {
	var req struct {
		Schema    string `json:"schema" binding:"required"`
		Module    string `json:"module" binding:"required"`
		Output    string `json:"output" binding:"required"`
		Detail    string `json:"detail"`
		ErrorInfo string `json:"error_info"`
		Service   string `json:"service"`
		ClientIP  string `json:"client_ip"`
		Operator  string `json:"operator"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据格式有误"})
		return
	}

	entry := &model.Log{
		LogBase: model.LogBase{
			Output:    req.Output,
			Detail:    req.Detail,
			ErrorInfo: req.ErrorInfo,
			Service:   req.Service,
			ClientIP:  req.ClientIP,
		},
		Schema:    model.LogSchema(req.Schema),
		Module:    model.LogModule(req.Module),
		PushType:  model.PushTypeHTTP,
		Timestamp: time.Now(),
	}
	if err := db.InsertLog(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "日志插入失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "日志已添加"})
}

func getLogs(c *gin.Context) {
	schema := c.Param("schema")
	module := c.Param("module")
	tableName := fmt.Sprintf("%s.%s%s_%s", schema, db.TablePrefix, schema, module)

	if err := db.EnsureTable(schema, module); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "表不存在或创建失败"})
		return
	}

	rows, err := db.ClickHouseDB.Query("SELECT * FROM " + tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询日志失败"})
		return
	}
	defer rows.Close()

	var logs []model.LogEntry
	for rows.Next() {
		var entry model.LogEntry
		if err := rows.Scan(&entry.Output, &entry.Detail, &entry.ErrorInfo, &entry.Service,
			&entry.ClientIP, &entry.ClientAddr, &entry.Operator, &entry.OperationTime); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "日志解析失败"})
			return
		}
		logs = append(logs, entry)
	}
	c.JSON(http.StatusOK, logs)
}
