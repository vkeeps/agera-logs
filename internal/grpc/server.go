package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vkeeps/agera-logs/internal/db"
	"github.com/vkeeps/agera-logs/internal/model"
	"github.com/vkeeps/agera-logs/proto"
	"google.golang.org/grpc/peer"
)

type LogServer struct {
	proto.UnimplementedLogServiceServer
	Logger *logrus.Logger
}

func (s *LogServer) SendLog(ctx context.Context, req *proto.LogRequest) (*proto.LogResponse, error) {
	clientIP, clientAddr := "0.0.0.0", "unknown"
	if p, ok := peer.FromContext(ctx); ok {
		clientIP, clientAddr = parseRemoteAddr(p.Addr.String())
	}

	// 检查 service 是否为空
	if req.Service == "" {
		s.Logger.Error(fmt.Sprintf("gRPC 日志缺少 service 字段，跳过插入，原始数据: %+v", req))
		return &proto.LogResponse{Success: false}, fmt.Errorf("service 字段为空")
	}

	schemaID := db.GenerateSchemaID(req.Schema)
	schemaName, err := db.GetSchemaNameByID(schemaID, s.Logger)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("获取 schema_id %s 失败: %v", schemaID, err))
		return &proto.LogResponse{Success: false}, err
	}
	if schemaName == "" {
		s.Logger.Warn(fmt.Sprintf("无效的 schema_id: %s，未在 BoltDB 中注册, schema: %s", schemaID, req.Schema))
		db.RebuildSchemaCache(schemaID, s.Logger)
		start := time.Now()
		timeout := 100 * time.Millisecond
		for time.Since(start) < timeout {
			schemaName, err = db.GetSchemaNameByID(schemaID, s.Logger)
			if err == nil && schemaName != "" {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if schemaName == "" {
			s.Logger.Error(fmt.Sprintf("重试后仍无效的 schema_id: %s，跳过插入", schemaID))
			return &proto.LogResponse{Success: false}, fmt.Errorf("无效的 schema_id: %s，未在 BoltDB 中注册", schemaID)
		}
	}

	entry := &model.Log{
		LogBase: model.LogBase{
			Output:     req.Output,
			Detail:     req.Detail,
			ErrorInfo:  req.ErrorInfo,
			Service:    req.Service,
			ClientIP:   clientIP,
			ClientAddr: clientAddr,
			LogLevel:   req.LogLevel,
		},
		Schema:            model.LogSchema(schemaName),
		Module:            model.LogModule(req.Module),
		PushType:          model.PushTypeGRPC,
		Timestamp:         time.Now(),
		OperatorID:        req.OperatorID,
		Operator:          req.Operator,
		OperatorIP:        req.OperatorIP,
		OperatorEquipment: req.OperatorEquipment,
		OperatorCompany:   req.OperatorCompany,
		OperatorProject:   req.OperatorProject,
	}

	if err := db.InsertLogs([]*model.Log{entry}, s.Logger); err != nil {
		s.Logger.Error(fmt.Sprintf("gRPC 日志插入失败: %v", err))
		return &proto.LogResponse{Success: false}, err
	}

	return &proto.LogResponse{Success: true}, nil
}

func parseRemoteAddr(addr string) (clientIP, clientAddr string) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "0.0.0.0", addr
	}
	return host, addr
}
