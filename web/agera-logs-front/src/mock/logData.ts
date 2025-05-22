import { LogEntry, LogLevel, LogFilterOptions } from '../types/logs'
import dayjs from 'dayjs'

// 服务列表
const services = [
  'api-gateway',
  'auth-service',
  'user-service',
  'payment-service',
  'notification-service',
  'recommendation-service',
  'analytics-service'
]

// 日志来源
const sources = [
  'kubernetes',
  'docker',
  'application',
  'system',
  'database'
]

// 随机消息模板
const messageTemplates = [
  'Service started successfully',
  'Request processed in {time}ms',
  'Connection established with {service}',
  'Failed to connect to {service} after {retries} retries',
  'Database query executed in {time}ms',
  'User {userId} authenticated successfully',
  'Permission denied for user {userId}',
  'Rate limit exceeded for IP {ipAddress}',
  'Cache miss for key {key}',
  'Memory usage high: {percentage}%',
  'CPU usage spike detected: {percentage}%',
  'Unexpected exception: {error}',
  'API rate limit reached for client {clientId}',
  'Slow database query detected: {query}',
  'Service health check failed'
]

// 错误消息模板
const errorMessages = [
  'NullPointerException in {class}.{method}()',
  'Connection timeout after {seconds} seconds',
  'OutOfMemoryError: Java heap space',
  'Database connection failed: {reason}',
  'Assertion failed: {condition}',
  'Invalid configuration: {setting}',
  'Access denied to resource {resource}',
  'File not found: {path}',
  'Invalid JSON format: {data}',
  'SSL handshake failed'
]

// 详细信息模板
const detailsTemplates = [
  {
    stackTrace: [
      'at com.example.{service}.{class}.{method}({file}:{line})',
      'at com.example.{service}.{class2}.{method2}({file2}:{line2})',
      'at java.base/java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1136)'
    ],
    environment: {
      javaVersion: '{javaVersion}',
      osName: '{osName}',
      memoryUsage: '{memoryUsage}MB'
    }
  },
  {
    request: {
      method: '{method}',
      path: '{path}',
      query: '{query}',
      headers: {
        'user-agent': '{userAgent}',
        'content-type': '{contentType}'
      }
    },
    response: {
      statusCode: '{statusCode}',
      responseTime: '{responseTime}ms'
    }
  },
  {
    metrics: {
      cpuUsage: '{cpuUsage}%',
      memoryUsage: '{memoryUsage}MB',
      diskUsage: '{diskUsage}%',
      threadCount: '{threadCount}'
    }
  }
]

// 随机数生成函数
const random = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1)) + min

// 随机选择数组中的一个元素
const randomChoice = <T>(arr: T[]): T => arr[Math.floor(Math.random() * arr.length)]

// 替换模板中的占位符
const fillTemplate = (template: string): string => {
  return template
    .replace('{time}', random(5, 500).toString())
    .replace('{service}', randomChoice(services))
    .replace('{retries}', random(1, 5).toString())
    .replace('{userId}', `user-${random(1000, 9999)}`)
    .replace('{ipAddress}', `192.168.${random(1, 255)}.${random(1, 255)}`)
    .replace('{key}', `cache:${random(1000, 9999)}`)
    .replace('{percentage}', random(50, 99).toString())
    .replace('{error}', 'java.lang.NullPointerException')
    .replace('{clientId}', `client-${random(100, 999)}`)
    .replace('{query}', 'SELECT * FROM users WHERE id = ?')
    .replace('{class}', `ServiceImpl`)
    .replace('{method}', `processRequest`)
    .replace('{seconds}', random(30, 120).toString())
    .replace('{reason}', 'Connection refused')
    .replace('{condition}', 'value != null')
    .replace('{setting}', 'database.url')
    .replace('{resource}', '/api/admin')
    .replace('{path}', '/etc/config/app.conf')
    .replace('{data}', '{"invalid": "json"')
}

// 生成详细信息
const generateDetails = (): Record<string, any> | undefined => {
  if (Math.random() > 0.7) {
    const template = JSON.parse(JSON.stringify(randomChoice(detailsTemplates)))
    
    // 递归遍历并填充模板
    const fillObject = (obj: any): any => {
      if (typeof obj === 'string') {
        return fillTemplate(obj)
      } else if (Array.isArray(obj)) {
        return obj.map(item => fillObject(item))
      } else if (obj !== null && typeof obj === 'object') {
        const result: Record<string, any> = {}
        for (const key in obj) {
          result[key] = fillObject(obj[key])
        }
        return result
      }
      return obj
    }
    
    return fillObject(template)
  }
  return undefined
}

// 生成随机日志条目
const generateRandomLog = (id: number, timeOffset = 0): LogEntry => {
  const level = randomChoice([
    LogLevel.DEBUG,
    LogLevel.INFO,
    LogLevel.INFO,
    LogLevel.INFO,
    LogLevel.WARNING,
    LogLevel.ERROR,
  ])
  
  const service = randomChoice(services)
  const source = randomChoice(sources)
  const traceId = `trace-${random(10000, 99999)}`
  
  // 根据级别选择合适的消息模板
  let message = ''
  if (level === LogLevel.ERROR) {
    message = fillTemplate(randomChoice(errorMessages))
  } else {
    message = fillTemplate(randomChoice(messageTemplates))
  }
  
  // 生成时间戳 - 越近的日志时间越近
  const now = dayjs()
  const timestamp = now.subtract(timeOffset, 'minutes').format('YYYY-MM-DD HH:mm:ss')
  
  return {
    id: `log-${id}`,
    timestamp,
    level,
    service,
    message,
    source,
    traceId,
    details: generateDetails()
  }
}

// 生成分页的日志数据
export const generateLogs = (count = 100): LogEntry[] => {
  const logs: LogEntry[] = []
  for (let i = 0; i < count; i++) {
    logs.push(generateRandomLog(i, i))
  }
  return logs
}

// 根据过滤条件筛选日志
export const filterLogs = (logs: LogEntry[], filters: LogFilterOptions): LogEntry[] => {
  return logs.filter(log => {
    // 级别过滤
    if (filters.level && log.level !== filters.level) {
      return false
    }
    
    // 服务过滤
    if (filters.service && log.service !== filters.service) {
      return false
    }
    
    // 来源过滤
    if (filters.source && log.source !== filters.source) {
      return false
    }
    
    // 时间范围过滤
    if (filters.timeRange && filters.timeRange.length === 2) {
      const logTime = dayjs(log.timestamp)
      const startTime = dayjs(filters.timeRange[0])
      const endTime = dayjs(filters.timeRange[1])
      
      if (logTime.isBefore(startTime) || logTime.isAfter(endTime)) {
        return false
      }
    }
    
    // 搜索文本过滤
    if (filters.searchText && !log.message.toLowerCase().includes(filters.searchText.toLowerCase())) {
      return false
    }
    
    return true
  })
}

// 获取所有服务名
export const getServices = (): string[] => {
  return [...services]
}

// 获取所有来源
export const getSources = (): string[] => {
  return [...sources]
} 