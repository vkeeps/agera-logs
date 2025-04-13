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
	Detail     string `json:"detail,omitempty"`
	ErrorInfo  string `json:"error_info,omitempty"`
	Service    string `json:"service,omitempty"`
	ClientIP   string `json:"client_ip,omitempty"`
	ClientAddr string `json:"client_addr,omitempty"`
	Operator   string `json:"operator,omitempty"`
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

		// 验证 schema_id 是否合法
		schemaName, err := db.GetSchemaNameByID(req.SchemaID)
		if err != nil {
			log.Printf("获取 schema_id %s 失败: %v", req.SchemaID, err)
			continue
		}
		if schemaName == "" {
			log.Printf("无效的 schema_id: %s，未在 BoltDB 中注册", req.SchemaID)
			continue
		}

		// 验证 module 是否合法
		if !validModule(req.Module) {
			log.Printf("无效的 module: %s", req.Module)
			continue
		}

		// 构造日志条目
		entry := &model.Log{
			LogBase: model.LogBase{
				Output:     req.Output,
				Detail:     req.Detail,
				ErrorInfo:  req.ErrorInfo,
				Service:    req.Service,
				ClientIP:   req.ClientIP,
				ClientAddr: req.ClientAddr,
			},
			Schema:    model.LogSchema(schemaName),
			Module:    model.LogModule(req.Module),
			PushType:  model.PushTypeUDP,
			Timestamp: time.Now(),
		}

		// 插入日志（表不存在会自动创建）
		if err := db.InsertLog(entry); err != nil {
			log.Printf("插入日志失败: %v", err)
		}
	}
}
