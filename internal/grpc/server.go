package grpc

import (
	"context"
	"log"
	"time"

	"github.com/vkeeps/agera-logs/internal/db"
	"github.com/vkeeps/agera-logs/internal/model"
	"github.com/vkeeps/agera-logs/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LogServer struct {
	proto.UnimplementedLogServiceServer
}

// validSchema 检查 schema 是否合法
func validSchema(schema string) bool {
	switch model.LogSchema(schema) {
	case model.SchemaLogin, model.SchemaAction:
		return true
	default:
		return false
	}
}

// validModule 检查 module 是否合法
func validModule(module string) bool {
	switch model.LogModule(module) {
	case model.ModuleLogin, model.ModuleLogout, model.ModuleError,
		model.ModulePermission, model.ModuleUser, model.ModuleGroup:
		return true
	default:
		return false
	}
}

func (s *LogServer) SendLog(ctx context.Context, req *proto.LogRequest) (*proto.LogResponse, error) {
	// 验证 schema
	if !validSchema(req.Schema) {
		log.Printf("无效的 schema: %s", req.Schema)
		return nil, status.Errorf(codes.InvalidArgument, "无效的 schema: %s", req.Schema)
	}

	// 验证 module
	if !validModule(req.Module) {
		log.Printf("无效的 module: %s", req.Module)
		return nil, status.Errorf(codes.InvalidArgument, "无效的 module: %s", req.Module)
	}

	// 构造日志条目
	entry := &model.Log{
		LogBase: model.LogBase{
			Output:     req.Output,
			Detail:     req.Detail,
			ErrorInfo:  req.ErrorInfo,
			Service:    req.Service,
			ClientIP:   req.ClientIp,
			ClientAddr: req.ClientAddr,
		},
		Schema:    model.LogSchema(req.Schema),
		Module:    model.LogModule(req.Module),
		PushType:  model.PushTypeGRPC,
		Timestamp: time.Now(),
	}

	// 插入日志（表不存在会自动创建）
	if err := db.InsertLog(entry); err != nil {
		log.Printf("日志插入失败: %v", err)
		return &proto.LogResponse{Success: false}, status.Errorf(codes.Internal, "日志插入失败: %v", err)
	}

	return &proto.LogResponse{Success: true}, nil
}
