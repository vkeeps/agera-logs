package main

import (
	"github.com/vkeeps/agera-logs/internal/db"
	"github.com/vkeeps/agera-logs/internal/grpc"
	"github.com/vkeeps/agera-logs/internal/http"
	"github.com/vkeeps/agera-logs/internal/udp"
	"github.com/vkeeps/agera-logs/proto"
	gg "google.golang.org/grpc"
	"log"
	"net"
	"os"
	"strconv"
)

func getAvailablePort(basePort int, protocol string) (net.Listener, int) {
	port := basePort
	for {
		if port > 65535 {
			log.Fatalf("无法找到可用的端口，端口号超出范围")
		}
		lis, err := net.Listen(protocol, ":"+strconv.Itoa(port))
		if err == nil {
			return lis, port
		}
		log.Printf("端口 %d 被占，试下个: %v", port, err)
		port++
	}
}

func main() {
	db.InitBolt()
	db.InitClickHouse()

	grpcBasePort := 50051
	grpcLis, grpcPort := getAvailablePort(grpcBasePort, "tcp")
	os.Setenv("GRPC_PORT", strconv.Itoa(grpcPort))
	s := gg.NewServer()
	proto.RegisterLogServiceServer(s, &grpc.LogServer{})
	go func() {
		log.Printf("gRPC 服务跑起来了，端口: %d", grpcPort)
		if err := s.Serve(grpcLis); err != nil {
			log.Fatalf("gRPC 服务挂了: %v", err)
		}
	}()

	udpBasePort := 50052
	go udp.StartUDPServer(udpBasePort)

	httpPort := 9302
	if port := os.Getenv("HTTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			httpPort = p
		}
	}
	r := http.SetupRouter()
	log.Printf("HTTP 服务跑起来了，端口: %d", httpPort)
	if err := r.Run(":" + strconv.Itoa(httpPort)); err != nil {
		log.Fatalf("HTTP 服务挂了: %v", err)
	}
}
