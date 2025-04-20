package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vkeeps/agera-logs/internal/db"
	"github.com/vkeeps/agera-logs/internal/model"
)

func SetupRouter(log *logrus.Logger) *gin.Engine {
	r := gin.New()

	// 使用 logrus 替换 Gin 默认日志
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			log.Info(fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %d",
				param.ClientIP,
				param.TimeStamp.Format("2006-01-02 15:04:05"),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency.Milliseconds()))
			return ""
		},
	}))

	r.Use(gin.Recovery())

	r.POST("/logs", createLog(log))
	r.GET("/logs/:schema/:module", getLogs(log))
	r.POST("/schemas", createSchema(log))
	r.GET("/schemas/:name", getSchema(log))

	return r
}

func createSchema(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.BindJSON(&req); err != nil {
			log.Error("数据格式有误")
			c.JSON(http.StatusBadRequest, gin.H{"error": "数据格式有误"})
			return
		}

		schemaID, err := db.GetOrCreateSchema(req.Name, log)
		if err != nil {
			log.Error("创建 schema 失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建 schema 失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"schema": req.Name, "id": schemaID})
	}
}

func getSchema(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		schemaName := c.Param("name")
		schemaID, err := db.GetOrCreateSchema(schemaName, log)
		if err != nil {
			log.Error("查询 schema 失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询 schema 失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"schema": schemaName, "id": schemaID})
	}
}

func createLog(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Schema     string `json:"schema" binding:"required"`
			Module     string `json:"module" binding:"required"`
			Output     string `json:"output" binding:"required"`
			Detail     string `json:"detail"`
			ErrorInfo  string `json:"error_info"`
			Service    string `json:"service"`
			ClientIP   string `json:"client_ip"`
			OperatorID string `json:"operator_id"`
			Operator   string `json:"operator"`
		}
		if err := c.BindJSON(&req); err != nil {
			log.Error("数据格式有误")
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
			Schema:     model.LogSchema(req.Schema),
			Module:     model.LogModule(req.Module),
			PushType:   model.PushTypeHTTP,
			Timestamp:  time.Now(),
			OperatorID: req.OperatorID, // 支持 operator_id
			Operator:   req.Operator,   // 支持 operator
		}

		if err := db.InsertLogs([]*model.Log{entry}, log); err != nil {
			log.Error("日志插入失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "日志插入失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "日志已添加"})
	}
}

func getLogs(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		schema := c.Param("schema")
		module := c.Param("module")
		tableName := fmt.Sprintf("%s.%s%s_%s", schema, db.TablePrefix, schema, module)

		if err := db.EnsureTable(schema, module, log); err != nil {
			log.Error("表不存在或创建失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "表不存在或创建失败"})
			return
		}

		rows, err := db.ClickHouseDB.Query("SELECT output, detail, error_info, service, client_ip, client_addr, operator_id, operator, operation_time, push_type FROM " + tableName)
		if err != nil {
			log.Error("查询日志失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询日志失败"})
			return
		}
		defer rows.Close()

		var logs []struct {
			Output        string
			Detail        string
			ErrorInfo     string
			Service       string
			ClientIP      string
			ClientAddr    string
			OperatorID    string
			Operator      string
			OperationTime time.Time
			PushType      string
		}
		for rows.Next() {
			var entry struct {
				Output        string
				Detail        string
				ErrorInfo     string
				Service       string
				ClientIP      string
				ClientAddr    string
				OperatorID    string
				Operator      string
				OperationTime time.Time
				PushType      string
			}
			if err := rows.Scan(&entry.Output, &entry.Detail, &entry.ErrorInfo, &entry.Service,
				&entry.ClientIP, &entry.ClientAddr, &entry.OperatorID, &entry.Operator,
				&entry.OperationTime, &entry.PushType); err != nil {
				log.Error("日志解析失败")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "日志解析失败"})
				return
			}
			logs = append(logs, entry)
		}
		c.JSON(http.StatusOK, logs)
	}
}
