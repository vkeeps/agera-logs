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
)

type Log struct {
	Schema     LogSchema
	Module     LogModule
	Output     string
	Detail     string
	ErrorInfo  string
	Service    string
	ClientIP   string
	ClientAddr string
	PushType   LogPushType
	Timestamp  time.Time
}
