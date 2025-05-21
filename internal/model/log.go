package model

import "time"

type LogSchema string

const (
	SchemaLogin  LogSchema = "login"
	SchemaAction LogSchema = "action"
)

type LogModule string

const (
	ModuleLogin      LogModule = "login"
	ModuleLogout     LogModule = "logout"
	ModuleError      LogModule = "error"
	ModulePermission LogModule = "permission"
	ModuleUser       LogModule = "user"
	ModuleGroup      LogModule = "group"
)

type LogPushType string

const (
	PushTypeGRPC LogPushType = "grpc"
	PushTypeUDP  LogPushType = "udp"
	PushTypeHTTP LogPushType = "http"
	PushTypeTCP  LogPushType = "tcp"
)

// LogBase 基础日志字段，供 Log 和 LogEntry 复用
type LogBase struct {
	Output     string `json:"output"`
	Detail     string `json:"detail"`
	ErrorInfo  string `json:"error_info"`
	Service    string `json:"service"`
	ClientIP   string `json:"client_ip"`
	ClientAddr string `json:"client_addr"`
	LogLevel   string `json:"log_level"`
}

// Log 推送时的完整日志模型
type Log struct {
	LogBase                       // 嵌入基础字段
	Schema            LogSchema   `json:"schema"`
	Module            LogModule   `json:"module"`
	PushType          LogPushType `json:"push_type"`
	Timestamp         time.Time   `json:"timestamp"`
	OperatorID        string      `json:"operator_id"`        // 操作人ID
	Operator          string      `json:"operator"`           // 操作人名称
	OperatorIP        string      `json:"operator_ip"`        // 扩展字段：操作人ip
	OperatorEquipment string      `json:"operator_equipment"` // 扩展字段：操作人的设备
	OperatorCompany   string      `json:"operator_company"`   // 扩展字段：操作人的企业
	OperatorProject   string      `json:"operator_project"`   // 扩展字段：操作人的项目（一个企业有多个项目这种）
}

// LogEntry 表中存储的日志模型
type LogEntry struct {
	LogBase                 // 嵌入基础字段
	OperatorID    string    `json:"operator_id"`    // 添加 operator_id
	Operator      string    `json:"operator"`       // 添加 operator
	OperationTime time.Time `json:"operation_time"` // 操作时间
}
