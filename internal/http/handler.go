package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vkeeps/agera-logs/internal/db"
	"github.com/vkeeps/agera-logs/internal/model"
)

func SetupRouter(log *logrus.Logger) *gin.Engine {
	r := gin.New()

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
	r.GET("/schemas", getAllSchemas(log))
	r.GET("/modules/:schemaId", getModulesBySchemaId(log))
	r.GET("/logs/by-schema/:schemaId", getLogsBySchemaId(log)) // 调整路由避免冲突

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

func getAllSchemas(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		schemas, err := db.GetAllSchemas(log)
		if err != nil {
			log.Error("查询所有 schema 失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询所有 schema 失败"})
			return
		}
		c.JSON(http.StatusOK, schemas)
	}
}

func getModulesBySchemaId(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		schemaId := c.Param("schemaId")
		modules, err := db.GetModulesBySchemaId(schemaId, log)
		if err != nil {
			log.Error("查询 schema 相关模块失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询 schema 相关模块失败"})
			return
		}
		c.JSON(http.StatusOK, modules)
	}
}

func createLog(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Schema            string `json:"schema" binding:"required"`
			Module            string `json:"module" binding:"required"`
			Output            string `json:"output" binding:"required"`
			Detail            string `json:"detail"`
			ErrorInfo         string `json:"error_info"`
			Service           string `json:"service"`
			ClientIP          string `json:"client_ip"`
			LogLevel          string `json:"log_level"`
			OperatorID        string `json:"operator_id"`
			Operator          string `json:"operator"`
			OperatorIP        string `json:"operator_ip"`
			OperatorEquipment string `json:"operator_equipment"`
			OperatorCompany   string `json:"operator_company"`
			OperatorProject   string `json:"operator_project"`
		}
		if err := c.BindJSON(&req); err != nil {
			log.Error("数据格式有误")
			c.JSON(http.StatusBadRequest, gin.H{"error": "数据格式有误"})
			return
		}

		// 检查 service 是否为空
		if req.Service == "" {
			log.Error(fmt.Sprintf("HTTP 日志缺少 service 字段，跳过插入，原始数据: %+v", req))
			c.JSON(http.StatusBadRequest, gin.H{"error": "service 字段为空"})
			return
		}

		entry := &model.Log{
			LogBase: model.LogBase{
				Output:    req.Output,
				Detail:    req.Detail,
				ErrorInfo: req.ErrorInfo,
				Service:   req.Service,
				ClientIP:  req.ClientIP,
				LogLevel:  req.LogLevel,
			},
			Schema:            model.LogSchema(req.Schema),
			Module:            model.LogModule(req.Module),
			PushType:          model.PushTypeHTTP,
			Timestamp:         time.Now(),
			OperatorID:        req.OperatorID,
			Operator:          req.Operator,
			OperatorIP:        req.OperatorIP,
			OperatorEquipment: req.OperatorEquipment,
			OperatorCompany:   req.OperatorCompany,
			OperatorProject:   req.OperatorProject,
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

		query := fmt.Sprintf("SELECT output, detail, error_info, service, client_ip, client_addr, log_level, operator_id, operator, operator_ip, operator_equipment, operator_company, operator_project, operation_time, push_type FROM %s ORDER BY operation_time DESC LIMIT 1000", tableName)
		rows, err := db.ClickHouseDB.Query(query)
		if err != nil {
			log.Error("查询日志失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询日志失败"})
			return
		}
		defer rows.Close()

		var logs []struct {
			Output            string
			Detail            string
			ErrorInfo         string
			Service           string
			ClientIP          string
			ClientAddr        string
			LogLevel          string
			OperatorID        string
			Operator          string
			OperatorIP        string
			OperatorEquipment string
			OperatorCompany   string
			OperatorProject   string
			OperationTime     time.Time
			PushType          string
		}
		for rows.Next() {
			var entry struct {
				Output            string
				Detail            string
				ErrorInfo         string
				Service           string
				ClientIP          string
				ClientAddr        string
				LogLevel          string
				OperatorID        string
				Operator          string
				OperatorIP        string
				OperatorEquipment string
				OperatorCompany   string
				OperatorProject   string
				OperationTime     time.Time
				PushType          string
			}
			if err := rows.Scan(&entry.Output, &entry.Detail, &entry.ErrorInfo, &entry.Service,
				&entry.ClientIP, &entry.ClientAddr, &entry.LogLevel, &entry.OperatorID, &entry.Operator,
				&entry.OperatorIP, &entry.OperatorEquipment, &entry.OperatorCompany, &entry.OperatorProject,
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

func getLogsBySchemaId(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		schemaId := c.Param("schemaId")

		// 先通过schemaId获取实际的数据库名称
		schemaName, err := db.GetSchemaNameByID(schemaId, log)
		if err != nil {
			log.Error(fmt.Sprintf("获取 schema_id %s 对应的数据库名失败: %v", schemaId, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取 schema_id %s 对应的数据库名失败", schemaId)})
			return
		}

		if schemaName == "" {
			log.Error(fmt.Sprintf("未找到 schema_id %s 对应的数据库名", schemaId))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("未找到 schema_id %s 对应的数据库名", schemaId)})
			return
		}

		// 使用实际的数据库名称查询表
		rows, err := db.ClickHouseDB.Query("SELECT name FROM system.tables WHERE database = ?", schemaName)
		if err != nil {
			log.Error(fmt.Sprintf("查询 schema %s 的表失败: %v", schemaName, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询 schema 相关表失败"})
			return
		}
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				log.Error(fmt.Sprintf("解析表名称失败: %v", err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "解析表名称失败"})
				return
			}
			if strings.HasPrefix(tableName, db.TablePrefix) {
				tables = append(tables, fmt.Sprintf("%s.%s", schemaName, tableName))
			}
		}

		var allLogs []struct {
			Module            string
			Output            string
			Detail            string
			ErrorInfo         string
			Service           string
			ClientIP          string
			ClientAddr        string
			LogLevel          string
			OperatorID        string
			Operator          string
			OperatorIP        string
			OperatorEquipment string
			OperatorCompany   string
			OperatorProject   string
			OperationTime     time.Time
			PushType          string
		}

		for _, table := range tables {
			query := fmt.Sprintf("SELECT output, detail, error_info, service, client_ip, client_addr, log_level, operator_id, operator, operator_ip, operator_equipment, operator_company, operator_project, operation_time, push_type FROM %s ORDER BY operation_time DESC LIMIT 1000", table)
			rows, err := db.ClickHouseDB.Query(query)
			if err != nil {
				log.Error("查询日志失败")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "查询日志失败"})
				return
			}
			defer rows.Close()

			for rows.Next() {
				var entry struct {
					Module            string
					Output            string
					Detail            string
					ErrorInfo         string
					Service           string
					ClientIP          string
					ClientAddr        string
					LogLevel          string
					OperatorID        string
					Operator          string
					OperatorIP        string
					OperatorEquipment string
					OperatorCompany   string
					OperatorProject   string
					OperationTime     time.Time
					PushType          string
				}
				if err := rows.Scan(&entry.Output, &entry.Detail, &entry.ErrorInfo, &entry.Service,
					&entry.ClientIP, &entry.ClientAddr, &entry.LogLevel, &entry.OperatorID, &entry.Operator,
					&entry.OperatorIP, &entry.OperatorEquipment, &entry.OperatorCompany, &entry.OperatorProject,
					&entry.OperationTime, &entry.PushType); err != nil {
					log.Error("日志解析失败")
					c.JSON(http.StatusInternalServerError, gin.H{"error": "日志解析失败"})
					return
				}
				entry.Module = table
				allLogs = append(allLogs, entry)
			}
		}

		for i := 0; i < len(allLogs)-1; i++ {
			for j := i + 1; j < len(allLogs); j++ {
				if allLogs[i].OperationTime.Before(allLogs[j].OperationTime) {
					allLogs[i], allLogs[j] = allLogs[j], allLogs[i]
				}
			}
		}

		c.JSON(http.StatusOK, allLogs)
	}
}
