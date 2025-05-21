package tcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vkeeps/agera-logs/internal/db"
	"github.com/vkeeps/agera-logs/internal/model"
)

type TCPLogRequest struct {
	SchemaID          string `json:"schema_id,schemaId"`
	Module            string `json:"module"`
	Output            string `json:"output"`
	Detail            string `json:"detail,omitempty"`
	ErrorInfo         string `json:"error_info,omitempty"`
	Service           string `json:"service,omitempty"`
	LogLevel          string `json:"log_level,omitempty"`
	OperatorID        string `json:"operator_id,operatorId"`
	Operator          string `json:"operator,omitempty"`
	OperatorIP        string `json:"operator_ip,omitempty"`
	OperatorEquipment string `json:"operator_equipment,omitempty"`
	OperatorCompany   string `json:"operator_company,omitempty"`
	OperatorProject   string `json:"operator_project,omitempty"`
}

var (
	batchSize      = 20
	batchTimeout   = 1 * time.Millisecond
	bufferCapacity = 500
)

var (
	insertedCount int64
	receivedCount int64
)

func init() {
	if size, err := strconv.Atoi(os.Getenv("BATCH_SIZE")); err == nil && size > 0 {
		batchSize = size
	}
	if timeout, err := time.ParseDuration(os.Getenv("BATCH_TIMEOUT")); err == nil && timeout > 0 {
		batchTimeout = timeout
	}
	if capacity, err := strconv.Atoi(os.Getenv("BUFFER_CAPACITY")); err == nil && capacity > 0 {
		bufferCapacity = capacity
	}
}

func validModule(module string) bool {
	return module != ""
}

func StartTCPServer(basePort int, stopChan chan struct{}, log *logrus.Logger) {
	port := basePort
	var listener *net.TCPListener
	for {
		if port > 65535 {
			log.Fatal("无法找到可用的 TCP 端口，端口号超出范围")
		}
		addr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			log.Error(fmt.Sprintf("TCP 地址解析失败: %v", err))
			port++
			continue
		}
		listener, err = net.ListenTCP("tcp", addr)
		if err != nil {
			log.Info(fmt.Sprintf("TCP 端口 %d 被占，试下个: %v", port, err))
			port++
			continue
		}
		break
	}
	defer listener.Close()

	os.Setenv("TCP_PORT", strconv.Itoa(port))
	log.Info(fmt.Sprintf("TCP 服务跑起来了，端口: %d", port))

	logBuffer := make([]*model.Log, 0, bufferCapacity)
	var mu sync.Mutex
	ticker := time.NewTicker(batchTimeout)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				if len(logBuffer) > 0 {
					entries := make([]*model.Log, len(logBuffer))
					copy(entries, logBuffer)
					logBuffer = logBuffer[:0]
					log.Info(fmt.Sprintf("准备插入 %d 条日志，缓冲区剩余 %d", len(entries), len(logBuffer)))
					mu.Unlock()

					start := time.Now()
					if err := db.InsertLogs(entries, log); err != nil {
						log.Error(fmt.Sprintf("批量插入 %d 条日志失败: %v", len(entries), err))
					} else {
						atomic.AddInt64(&insertedCount, int64(len(entries)))
						log.Info(fmt.Sprintf("成功插入 %d 条日志，耗时 %v，总计插入 %d 条", len(entries), time.Since(start), atomic.LoadInt64(&insertedCount)))
					}
				} else {
					mu.Unlock()
				}
			case <-stopChan:
				mu.Lock()
				if len(logBuffer) > 0 {
					entries := make([]*model.Log, len(logBuffer))
					copy(entries, logBuffer)
					logBuffer = logBuffer[:0]
					log.Info(fmt.Sprintf("停止服务，插入剩余 %d 条日志", len(entries)))
					mu.Unlock()
					if err := db.InsertLogs(entries, log); err != nil {
						log.Error(fmt.Sprintf("批量插入 %d 条日志失败: %v", len(entries), err))
					} else {
						atomic.AddInt64(&insertedCount, int64(len(entries)))
						log.Info(fmt.Sprintf("成功插入 %d 条日志，总计插入 %d 条", len(entries), atomic.LoadInt64(&insertedCount)))
					}
				} else {
					mu.Unlock()
				}
				return
			}
		}
	}()

	readTimeout := 1 * time.Second
	if t, err := time.ParseDuration(os.Getenv("READ_TIMEOUT")); err == nil && t > 0 {
		readTimeout = t
	}

	for {
		select {
		case <-stopChan:
			log.Info("收到停止信号，关闭 TCP 服务")
			return
		default:
			listener.SetDeadline(time.Now().Add(readTimeout))
			conn, err := listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Error(fmt.Sprintf("接受 TCP 连接失败: %v", err))
				continue
			}
			go handleConnection(conn, &logBuffer, &mu, stopChan, log)
		}
	}
}

func handleConnection(conn net.Conn, logBuffer *[]*model.Log, mu *sync.Mutex, stopChan chan struct{}, log *logrus.Logger) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	remoteAddr := conn.RemoteAddr().String()
	clientIP, clientAddr := parseRemoteAddr(remoteAddr)

	readTimeout := 1 * time.Second
	if t, err := time.ParseDuration(os.Getenv("READ_TIMEOUT")); err == nil && t > 0 {
		readTimeout = t
	}

	for {
		select {
		case <-stopChan:
			log.Info(fmt.Sprintf("停止处理连接: %s", remoteAddr))
			return
		default:
			conn.SetReadDeadline(time.Now().Add(readTimeout))
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err.Error() != "EOF" && !netErrTimeout(err) {
					log.Error(fmt.Sprintf("读取 TCP 数据失败: %v", err))
				}
				return
			}

			var req TCPLogRequest
			if err := json.Unmarshal(line, &req); err != nil {
				log.Error(fmt.Sprintf("TCP 数据解析失败: %v, 原始数据: %s", err, string(line)))
				continue
			}

			if req.SchemaID == "" {
				log.Error(fmt.Sprintf("收到空 schema_id，原始数据: %s", string(line)))
				continue
			}

			// 检查 service 是否为空
			if req.Service == "" {
				log.Error(fmt.Sprintf("TCP 日志缺少 service 字段，跳过插入，原始数据: %+v", req))
				continue
			}

			schemaName, err := db.GetSchemaNameByID(req.SchemaID, log)
			if err != nil {
				log.Error(fmt.Sprintf("获取 schema_id %s 失败: %v", req.SchemaID, err))
				continue
			}
			if schemaName == "" {
				log.Warn(fmt.Sprintf("无效的 schema_id: %s，未在 BoltDB 中注册, 原始数据: %s", req.SchemaID, string(line)))
				db.RebuildSchemaCache(req.SchemaID, log)
				start := time.Now()
				timeout := 100 * time.Millisecond
				for time.Since(start) < timeout {
					schemaName, err = db.GetSchemaNameByID(req.SchemaID, log)
					if err == nil && schemaName != "" {
						break
					}
					time.Sleep(10 * time.Millisecond)
				}
				if schemaName == "" {
					log.Error(fmt.Sprintf("重试后仍无效的 schema_id: %s，跳过插入", req.SchemaID))
					continue
				}
			}

			if !validModule(req.Module) {
				log.Error(fmt.Sprintf("无效的 module: %s", req.Module))
				continue
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
				PushType:          model.PushTypeTCP,
				Timestamp:         time.Now(),
				OperatorID:        req.OperatorID,
				Operator:          req.Operator,
				OperatorIP:        req.OperatorIP,
				OperatorEquipment: req.OperatorEquipment,
				OperatorCompany:   req.OperatorCompany,
				OperatorProject:   req.OperatorProject,
			}

			mu.Lock()
			if len(*logBuffer) < bufferCapacity {
				*logBuffer = append(*logBuffer, entry)
				count := atomic.AddInt64(&receivedCount, 1)
				log.Info(fmt.Sprintf("收到第 %d 条数据，从 %s", count, remoteAddr))
			} else {
				log.Error(fmt.Sprintf("缓冲区已满，丢弃日志: %v", entry))
			}
			if len(*logBuffer) >= batchSize {
				entries := make([]*model.Log, len(*logBuffer))
				copy(entries, *logBuffer)
				*logBuffer = (*logBuffer)[:0]
				log.Info(fmt.Sprintf("准备插入 %d 条日志，缓冲区剩余 %d", len(entries), len(*logBuffer)))
				mu.Unlock()

				start := time.Now()
				if err := db.InsertLogs(entries, log); err != nil {
					log.Error(fmt.Sprintf("批量插入 %d 条日志失败: %v", len(entries), err))
				} else {
					atomic.AddInt64(&insertedCount, int64(len(entries)))
					log.Info(fmt.Sprintf("成功插入 %d 条日志，耗时 %v，总计插入 %d 条", len(entries), time.Since(start), atomic.LoadInt64(&insertedCount)))
				}
			} else {
				mu.Unlock()
			}
		}
	}
}

func parseRemoteAddr(addr string) (clientIP, clientAddr string) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "0.0.0.0", addr
	}
	return host, addr
}

func netErrTimeout(err error) bool {
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	return false
}
