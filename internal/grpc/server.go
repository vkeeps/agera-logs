package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vkeeps/agera-logs/internal/db"
	"github.com/vkeeps/agera-logs/internal/model"
	"github.com/vkeeps/agera-logs/proto"
)

type LogServer struct {
	proto.UnimplementedLogServiceServer
	Logger *logrus.Logger
}

func (s *LogServer) SendLog(ctx context.Context, req *proto.LogRequest) (*proto.LogResponse, error) {
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

	if err := db.InsertLogs([]*model.Log{entry}, s.Logger); err != nil {
		s.Logger.Error(fmt.Sprintf("gRPC 日志插入失败: %v", err))
		return &proto.LogResponse{Success: false}, err
	}

	return &proto.LogResponse{Success: true}, nil
}
