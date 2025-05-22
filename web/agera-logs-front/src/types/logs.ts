// 日志级别枚举
export enum LogLevel {
  DEBUG = 'debug',
  INFO = 'info',
  WARNING = 'warning',
  ERROR = 'error'
}

// 后端日志条目接口
export interface BackendLogEntry {
  Output: string
  Detail: string
  ErrorInfo: string
  Service: string
  ClientIP: string
  ClientAddr: string
  LogLevel: string
  OperatorID: string
  Operator: string
  OperatorIP: string
  OperatorEquipment: string
  OperatorCompany: string
  OperatorProject: string
  OperationTime: string
  PushType: string
  Module?: string // 模块信息，某些查询会返回
}

// 前端使用的日志条目接口
export interface LogEntry {
  id: string
  timestamp: string
  level: LogLevel
  service: string
  message: string
  details?: Record<string, any>
  source?: string
  traceId?: string
  module?: string
  operator?: string
  operatorInfo?: {
    id: string
    ip: string
    equipment: string
    company: string
    project: string
  }
}

// 日志过滤选项
export interface LogFilterOptions {
  level?: LogLevel | ''
  service?: string
  timeRange?: [string, string] | null
  searchText?: string
  source?: string
  schema?: string
  module?: string
  operatorId?: string
}

// 分页接口
export interface Pagination {
  current: number
  pageSize: number
  total: number
}

// Schema 接口
export interface Schema {
  id: string
  name: string
}

// Module 接口
export interface Module {
  id: string
  name: string
  schema_id: string
}

// 创建日志请求参数
export interface CreateLogRequest {
  schema: string
  module: string
  output: string
  detail?: string
  error_info?: string
  service: string
  client_ip?: string
  log_level?: string
  operator_id?: string
  operator?: string
  operator_ip?: string
  operator_equipment?: string
  operator_company?: string
  operator_project?: string
} 