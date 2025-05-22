import request from './request'
import { 
  BackendLogEntry, 
  LogEntry, 
  LogLevel, 
  Schema, 
  Module, 
  CreateLogRequest 
} from '../types/logs'

// 获取所有schema
export async function getAllSchemas(): Promise<Schema[]> {
  try {
    // 直接调用schemas接口
    return await request.get('/schemas')
  } catch (error) {
    console.error('获取所有schemas失败:', error)
    return []
  }
}

// 根据名称获取schema
export async function getSchemaByName(name: string): Promise<Schema> {
  try {
    return await request.get(`/schemas/${name}`)
  } catch (error) {
    console.error(`获取schema ${name}失败:`, error)
    return { id: '', name: '' }
  }
}

// 创建一个新的schema
export async function createSchema(name: string): Promise<Schema> {
  try {
    return await request.post('/schemas', { name })
  } catch (error) {
    console.error('创建schema失败:', error)
    throw error // 创建失败应当抛出错误
  }
}

// 根据schemaId获取模块列表
export async function getModulesBySchemaId(schemaId: string): Promise<Module[]> {
  try {
    // 正确调用modules/:schemaId接口
    console.log(`获取schemaId ${schemaId}的模块列表`)
    const result = await request.get(`/modules/${schemaId}`)
    console.log(`获取到模块列表结果:`, result)
    
    // 后端可能返回空数据
    if (!result) {
      console.warn(`获取模块列表失败，返回为空`)
      return []
    }

    // 处理后端返回的字符串数组 ["MODULE_NAME"]
    if (Array.isArray(result)) {
      if (result.length > 0 && typeof result[0] === 'string') {
        console.log('将字符串模块名转换为Module对象')
        return result.map((moduleName, index) => ({
          id: `module_${index}`,
          name: moduleName,
          schema_id: schemaId
        }))
      }
      
      if (result.length > 0 && typeof result[0] === 'object') {
        // 已经是Module对象数组
        return result as Module[]
      }

      return []
    } 
    
    // 如果是其他格式的对象，尝试解析
    if (typeof result === 'object' && result !== null) {
      const moduleData = result as any
      
      // 检查是否有modules字段
      if (moduleData.modules && Array.isArray(moduleData.modules)) {
        return moduleData.modules
      }
    }
    
    console.warn('未知的模块列表格式，返回空数组:', result)
    return []
  } catch (error) {
    console.error(`获取schemaId ${schemaId}的模块失败:`, error)
    return []
  }
}

// 根据schema名称和module名称获取日志
export async function getLogsBySchemaAndModule(schema: string, module: string): Promise<LogEntry[]> {
  try {
    // 直接使用schema和module名称，而不是ID
    const response = await request.get(`/logs/${schema}/${module}`)
    return transformLogs(response)
  } catch (error) {
    console.error(`获取${schema}.${module}的日志失败:`, error)
    return []
  }
}

// 根据schemaId获取所有日志
export async function getLogsBySchemaId(schemaId: string): Promise<LogEntry[]> {
  try {
    // 使用正确的by-schema路径参数
    const response = await request.get(`/logs/by-schema/${schemaId}`)
    return transformLogs(response)
  } catch (error) {
    console.error(`获取schemaId ${schemaId}的所有日志失败:`, error)
    return []
  }
}

// 创建日志
export async function createLog(logData: CreateLogRequest): Promise<any> {
  try {
    return await request.post('/logs', logData)
  } catch (error) {
    console.error('创建日志失败:', error)
    throw error
  }
}

// 将后端日志格式转换为前端格式
function transformLogs(backendLogs: any[]): LogEntry[] {
  // 如果后端返回的数据为空或者不是数组，则返回空数组
  if (!backendLogs || !Array.isArray(backendLogs)) {
    return []
  }

  return backendLogs.map((log, index) => {
    // 确定日志级别
    let level: LogLevel
    switch (log?.LogLevel?.toLowerCase() || log?.log_level?.toLowerCase()) {
      case 'debug':
        level = LogLevel.DEBUG
        break
      case 'info':
        level = LogLevel.INFO
        break
      case 'warning':
      case 'warn':
        level = LogLevel.WARNING
        break
      case 'error':
      case 'fatal':
        level = LogLevel.ERROR
        break
      default:
        level = LogLevel.INFO
    }
    
    // 构建详细信息对象
    let details: Record<string, any> = {}
    
    if (log?.Detail || log?.detail) {
      try {
        // 尝试解析JSON字符串
        details = JSON.parse(log?.Detail || log?.detail)
      } catch (e) {
        // 如果不是JSON，就直接作为文本
        details = { text: log?.Detail || log?.detail }
      }
    }
    
    // 如果有错误信息，添加到详细信息中
    if (log?.ErrorInfo || log?.error_info) {
      details.errorInfo = log?.ErrorInfo || log?.error_info
    }

    // 提取操作时间
    const timestamp = log?.OperationTime || log?.operation_time || new Date().toISOString()
    
    return {
      id: `log-${index}-${Date.now()}`, // 生成一个唯一ID
      timestamp: typeof timestamp === 'string' ? timestamp : timestamp.toISOString(),
      level,
      service: log?.Service || log?.service || '未知服务',
      message: log?.Output || log?.output || '无消息内容',
      details: Object.keys(details).length > 0 ? details : undefined,
      source: log?.ClientAddr || log?.client_addr || log?.ClientIP || log?.client_ip || '未知来源',
      traceId: undefined, // 后端数据没有提供traceId
      module: log?.Module || log?.module || '未知模块',
      operator: log?.Operator || log?.operator || '',
      operatorInfo: {
        id: log?.OperatorID || log?.operator_id || '',
        ip: log?.OperatorIP || log?.operator_ip || '',
        equipment: log?.OperatorEquipment || log?.operator_equipment || '',
        company: log?.OperatorCompany || log?.operator_company || '',
        project: log?.OperatorProject || log?.operator_project || ''
      }
    }
  })
} 