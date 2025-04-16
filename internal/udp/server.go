package udp

import (
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

var (
	batchSize      = 20
	batchTimeout   = 1 * time.Millisecond
	bufferCapacity = 500
	ackPort        = 50054
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
	switch model.LogModule(module) {
	case model.ModuleLogin, model.ModuleLogout, model.ModuleError,
		model.ModulePermission, model.ModuleUser, model.ModuleGroup:
		return true
	default:
		return false
	}
}

func StartUDPServer(basePort int, stopChan chan struct{}, log *logrus.Logger) {
	port := basePort
	var conn *net.UDPConn
	for {
		if port > 65535 {
			log.Fatal("无法找到可用的 UDP 端口，端口号超出范围")
		}
		udpAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
		if err != nil {
			log.Error(fmt.Sprintf("UDP 地址解析失败: %v", err))
			port++
			continue
		}
		conn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			log.Info(fmt.Sprintf("UDP 端口 %d 被占，试下个: %v", port, err))
			port++
			continue
		}
		if err := conn.SetReadBuffer(2 * 1024 * 1024); err != nil {
			log.Error(fmt.Sprintf("设置 UDP 接收缓冲区失败: %v", err))
		}
		break
	}
	defer conn.Close()

	os.Setenv("UDP_PORT", strconv.Itoa(port))
	log.Info(fmt.Sprintf("UDP 服务跑起来了，端口: %d", port))

	go startAckServer(ackPort, stopChan, log)

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

	dataChan := make(chan struct {
		data []byte
		addr *net.UDPAddr
	}, bufferCapacity)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pkt := range dataChan {
				var req UDPLogRequest
				if err := json.Unmarshal(pkt.data, &req); err != nil {
					log.Error(fmt.Sprintf("UDP 数据解析失败: %v", err))
					continue
				}

				schemaName, err := db.GetSchemaNameByID(req.SchemaID, log)
				if err != nil {
					log.Error(fmt.Sprintf("获取 schema_id %s 失败: %v", req.SchemaID, err))
					continue
				}
				if schemaName == "" {
					log.Error(fmt.Sprintf("无效的 schema_id: %s，未在 BoltDB 中注册", req.SchemaID))
					continue
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
						ClientIP:   req.ClientIP,
						ClientAddr: req.ClientAddr,
					},
					Schema:    model.LogSchema(schemaName),
					Module:    model.LogModule(req.Module),
					PushType:  model.PushTypeUDP,
					Timestamp: time.Now(),
				}

				mu.Lock()
				if len(logBuffer) < bufferCapacity {
					logBuffer = append(logBuffer, entry)
					count := atomic.AddInt64(&receivedCount, 1)
					log.Info(fmt.Sprintf("收到第 %d 条数据，从 %s", count, pkt.addr.String()))
					go sendAck(pkt.addr, log)
				} else {
					log.Error(fmt.Sprintf("缓冲区已满，丢弃日志: %v", entry))
				}
				if len(logBuffer) >= batchSize {
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
			}
		}()
	}

	defer func() {
		close(dataChan)
		wg.Wait()
	}()

	readTimeout := 1 * time.Second
	if t, err := time.ParseDuration(os.Getenv("READ_TIMEOUT")); err == nil && t > 0 {
		readTimeout = t
	}

	buf := make([]byte, 4096)
	for {
		select {
		case <-stopChan:
			log.Info("收到停止信号，关闭 UDP 服务")
			return
		default:
			conn.SetReadDeadline(time.Now().Add(readTimeout))
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Error(fmt.Sprintf("UDP 读失败: %v", err))
				continue
			}
			data := make([]byte, n)
			copy(data, buf[:n])
			select {
			case dataChan <- struct {
				data []byte
				addr *net.UDPAddr
			}{data, addr}:
			default:
				log.Error("数据通道已满，丢弃数据")
			}
		}
	}
}

func startAckServer(port int, stopChan chan struct{}, log *logrus.Logger) {
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(fmt.Sprintf("解析确认端口 %d 失败: %v", port, err))
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(fmt.Sprintf("监听确认端口 %d 失败: %v", port, err))
	}
	defer conn.Close()
	log.Info(fmt.Sprintf("确认服务器跑起来了，端口: %d", port))

	readTimeout := 1 * time.Second
	if t, err := time.ParseDuration(os.Getenv("READ_TIMEOUT")); err == nil && t > 0 {
		readTimeout = t
	}

	buf := make([]byte, 1024)
	for {
		select {
		case <-stopChan:
			log.Info("停止确认服务器")
			return
		default:
			conn.SetReadDeadline(time.Now().Add(readTimeout))
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Error(fmt.Sprintf("读取确认数据失败: %v", err))
				continue
			}
			if string(buf[:n]) == "ACK_REQUEST" {
				_, err = conn.WriteToUDP([]byte("ACK"), addr)
				if err != nil {
					log.Error(fmt.Sprintf("发送确认失败: %v", err))
				}
			}
		}
	}
}

func sendAck(addr *net.UDPAddr, log *logrus.Logger) {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: addr.IP, Port: ackPort})
	if err != nil {
		log.Error(fmt.Sprintf("连接确认端口失败: %v", err))
		return
	}
	defer conn.Close()

	ackTimeout := 200 * time.Millisecond
	for retries := 3; retries > 0; retries-- {
		conn.SetWriteDeadline(time.Now().Add(ackTimeout))
		_, err = conn.Write([]byte("ACK_REQUEST"))
		if err != nil {
			log.Error(fmt.Sprintf("发送确认请求失败（剩余 %d 次重试）: %v", retries-1, err))
			time.Sleep(50 * time.Millisecond)
			continue
		}
		break
	}
}
