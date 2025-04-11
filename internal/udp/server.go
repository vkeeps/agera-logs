package udp

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/vkeeps/agera-logs/internal/db"
	"github.com/vkeeps/agera-logs/internal/model"
)

type UDPLogRequest struct {
	SchemaID   string `json:"schema_id"`
	Module     string `json:"module"`
	Output     string `json:"output"`
	Detail     string `json:"detail"`
	ErrorInfo  string `json:"error_info"`
	Service    string `json:"service"`
	ClientIP   string `json:"client_ip"`
	ClientAddr string `json:"client_addr"`
}

// StartUDPServer 启动 UDP 服务
func StartUDPServer(basePort int) {
	port := basePort
	var conn *net.UDPConn
	for {
		if port > 65535 {
			log.Fatalf("无法找到可用的 UDP 端口，端口号超出范围")
		}
		udpAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
		if err != nil {
			log.Printf("UDP 地址解析失败: %v", err)
			port++
			continue
		}
		conn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			log.Printf("UDP 端口 %d 被占，试下个: %v", port, err)
			port++
			continue
		}
		break
	}
	defer conn.Close()

	os.Setenv("UDP_PORT", strconv.Itoa(port))
	log.Printf("UDP 服务跑起来了，端口: %d", port)
	buf := make([]byte, 4096)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("UDP 读失败: %v", err)
			continue
		}

		var req UDPLogRequest
		if err := json.Unmarshal(buf[:n], &req); err != nil {
			log.Printf("UDP 数据解析失败: %v", err)
			continue
		}

		// 根据 schemaID 获取 schema 名称
		schemaName, err := db.GetSchemaNameByID(req.SchemaID)
		if err != nil || schemaName == "" {
			log.Printf("无效的 schema_id: %s", req.SchemaID)
			continue
		}

		entry := &model.Log{
			Schema:     model.LogSchema(schemaName),
			Module:     model.LogModule(req.Module),
			Output:     req.Output,
			Detail:     req.Detail,
			ErrorInfo:  req.ErrorInfo,
			Service:    req.Service,
			ClientIP:   req.ClientIP,
			ClientAddr: req.ClientAddr,
			PushType:   model.PushTypeUDP,
			Timestamp:  time.Now(),
		}
		if err := db.InsertLog(entry); err != nil {
			log.Printf("插入日志失败: %v", err)
		}
	}
}
