package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vkeeps/agera-logs/internal/db"
	"github.com/vkeeps/agera-logs/internal/grpc"
	"github.com/vkeeps/agera-logs/internal/http"
	"github.com/vkeeps/agera-logs/internal/logger"
	"github.com/vkeeps/agera-logs/internal/tcp"
	"github.com/vkeeps/agera-logs/internal/udp"
	"github.com/vkeeps/agera-logs/proto"
	gg "google.golang.org/grpc"
)

func getAvailablePort(basePort int, protocol string, log *logrus.Logger) (net.Listener, int, error) {
	port := basePort
	for {
		if port > 65535 {
			err := fmt.Errorf("无法找到可用的端口，端口号超出范围")
			log.Error(err.Error())
			return nil, 0, err
		}
		lis, err := net.Listen(protocol, ":"+strconv.Itoa(port))
		if err == nil {
			return lis, port, nil
		}
		log.Info(fmt.Sprintf("端口 %d 被占，试下个: %v", port, err))
		port++
	}
}

func main() {
	// 初始化日志
	log, err := logger.InitLogger("debug", "agera.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "日志初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化数据库
	db.InitBolt(log)
	db.InitClickHouse(log)

	// 上下文和信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 等待组确保服务启动和关闭
	var wg sync.WaitGroup

	// gRPC 服务
	grpcBasePort := 50051
	grpcLis, grpcPort, err := getAvailablePort(grpcBasePort, "tcp", log)
	if err != nil {
		log.Fatal(fmt.Sprintf("获取 gRPC 端口失败: %v", err))
	}
	os.Setenv("GRPC_PORT", strconv.Itoa(grpcPort))
	grpcServer := gg.NewServer()
	proto.RegisterLogServiceServer(grpcServer, &grpc.LogServer{Logger: log})
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info(fmt.Sprintf("gRPC 服务跑起来了，端口: %d", grpcPort))
		if err := grpcServer.Serve(grpcLis); err != nil && err != gg.ErrServerStopped {
			log.Error(fmt.Sprintf("gRPC 服务挂了: %v", err))
		}
	}()

	// UDP 服务
	udpBasePort := 50052
	udpStopChan := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info(fmt.Sprintf("启动 UDP 服务，基础端口: %d", udpBasePort))
		udp.StartUDPServer(udpBasePort, udpStopChan, log)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			close(udpStopChan)
		}
	}()

	// TCP 服务
	tcpBasePort := 50053
	tcpStopChan := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info(fmt.Sprintf("启动 TCP 服务，基础端口: %d", tcpBasePort))
		tcp.StartTCPServer(tcpBasePort, tcpStopChan, log)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			close(tcpStopChan)
		}
	}()

	// HTTP 服务（用 Gin）
	httpPort := 9302
	if port := os.Getenv("HTTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			httpPort = p
		}
	}
	_, httpPort, err = getAvailablePort(httpPort, "tcp", log)
	if err != nil {
		log.Fatal(fmt.Sprintf("获取 HTTP 端口失败: %v", err))
	}
	r := http.SetupRouter(log)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info(fmt.Sprintf("HTTP 服务跑起来了，端口: %d", httpPort))
		if err := r.Run(":" + strconv.Itoa(httpPort)); err != nil {
			log.Error(fmt.Sprintf("HTTP 服务挂了: %v", err))
		}
	}()

	// 等待信号
	go func() {
		<-sigChan
		log.Info("收到停止信号，开始关闭服务")
		cancel()

		// 关闭 gRPC
		grpcServer.GracefulStop()

		// 等待服务关闭
		shutdownTimeout := 5 * time.Second
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			log.Info("所有服务已优雅关闭")
		case <-shutdownCtx.Done():
			log.Error("服务关闭超时，强制退出")
		}
		os.Exit(0)
	}()

	// 主线程等待
	select {}
}
