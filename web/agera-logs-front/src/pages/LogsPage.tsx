import { useState, useEffect } from 'react'
import { Typography, Card, Alert, Spin, Pagination, Select, Form, Button } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import LogFilter from '../components/logs/LogFilter'
import LogTable from '../components/logs/LogTable'
import { LogEntry, LogFilterOptions, Pagination as PaginationType, Schema, Module } from '../types/logs'
import { 
  getAllSchemas, 
  getModulesBySchemaId, 
  getLogsBySchemaAndModule, 
  getLogsBySchemaId 
} from '../api/logService'

const { Title } = Typography
const { Option } = Select

const DEFAULT_PAGE_SIZE = 10

function LogsPage() {
  // 状态
  const [allLogs, setAllLogs] = useState<LogEntry[]>([])
  const [filteredLogs, setFilteredLogs] = useState<LogEntry[]>([])
  const [displayLogs, setDisplayLogs] = useState<LogEntry[]>([])
  const [pagination, setPagination] = useState<PaginationType>({
    current: 1,
    pageSize: DEFAULT_PAGE_SIZE,
    total: 0
  })
  
  // Schema和Module状态
  const [schemas, setSchemas] = useState<Schema[]>([])
  const [modules, setModules] = useState<Module[]>([])
  const [selectedSchema, setSelectedSchema] = useState<string>('')
  const [selectedModule, setSelectedModule] = useState<string>('')
  
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [services, setServices] = useState<string[]>([])
  const [sources, setSources] = useState<string[]>([])
  const [currentFilters, setCurrentFilters] = useState<LogFilterOptions>({})

  // 加载Schema列表
  useEffect(() => {
    const fetchSchemas = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await getAllSchemas()
        setSchemas(data)
        // 如果有Schema，默认选择第一个
        if (data.length > 0) {
          setSelectedSchema(data[0].id)
        }
        setLoading(false)
      } catch (err) {
        setError('后端服务不可用或无法加载Schema列表，请确认服务是否正常运行')
        console.error('加载Schema失败:', err)
        setLoading(false)
        // 清空数据
        setSchemas([])
        setModules([])
        setAllLogs([])
        setFilteredLogs([])
        setDisplayLogs([])
      }
    }
    
    fetchSchemas()
  }, [])
  
  // 当选择Schema变化时，加载对应的模块列表
  useEffect(() => {
    if (!selectedSchema) return
    
    const fetchModules = async () => {
      try {
        setLoading(true)
        setError(null)
        console.log(`正在获取schemaId ${selectedSchema}的模块列表`)
        const data = await getModulesBySchemaId(selectedSchema)
        console.log(`获取到模块列表:`, data)
        
        if (!data || data.length === 0) {
          console.warn('未获取到任何模块')
          setModules([])
          setSelectedModule('')
          setLoading(false)
          return
        }
        
        setModules(data)
        
        // 如果有模块，默认选择第一个
        console.log(`默认选择第一个模块: ${data[0].name}`)
        setSelectedModule(data[0].name)
        setLoading(false)
      } catch (err) {
        setError('加载模块列表失败')
        console.error('加载模块失败:', err)
        setLoading(false)
        setModules([])
        setSelectedModule('')
      }
    }
    
    fetchModules()
  }, [selectedSchema])
  
  // 当Schema或Module变化时，加载日志数据
  useEffect(() => {
    if (!selectedSchema) return
    
    const fetchLogs = async () => {
      setLoading(true)
      setError(null)
      
      try {
        let logs: LogEntry[] = []
        
        // 获取选中schema的信息（需要schema名称）
        const selectedSchemaInfo = schemas.find(s => s.id === selectedSchema)
        
        if (!selectedSchemaInfo) {
          setError('无法找到所选Schema的信息')
          setLoading(false)
          return
        }
        
        // 获取日志数据
        if (selectedModule && selectedModule.trim() !== '') {
          // 如果选择了特定模块，则获取该模块的日志
          console.log(`正在获取 ${selectedSchemaInfo.name}.${selectedModule} 的日志`)
          logs = await getLogsBySchemaAndModule(selectedSchemaInfo.name, selectedModule) || []
        } else {
          // 没有选择模块时，如果有模块列表，则获取第一个模块的日志
          if (modules.length > 0) {
            const firstModule = modules[0].name
            console.log(`没有选择模块，自动获取第一个模块 ${selectedSchemaInfo.name}.${firstModule} 的日志`)
            logs = await getLogsBySchemaAndModule(selectedSchemaInfo.name, firstModule) || []
          } else {
            // 如果没有模块列表，尝试获取Schema下的所有日志
            console.log(`正在获取schemaId ${selectedSchema} 的所有日志`)
            logs = await getLogsBySchemaId(selectedSchema) || []
          }
        }
        
        // 确保logs非空
        logs = logs || []
        console.log(`获取到 ${logs.length} 条日志记录`)
        
        setAllLogs(logs)
        setFilteredLogs(logs)
        setPagination({
          ...pagination,
          current: 1,
          total: logs.length
        })
        
        // 提取所有服务和来源，用于过滤
        const uniqueServices = new Set<string>()
        const uniqueSources = new Set<string>()
        
        logs.forEach(log => {
          if (log.service) uniqueServices.add(log.service)
          if (log.source) uniqueSources.add(log.source)
        })
        
        setServices(Array.from(uniqueServices))
        setSources(Array.from(uniqueSources))
        
        // 更新显示的日志
        updateDisplayLogs(logs, 1, pagination.pageSize)
        setLoading(false)
      } catch (err) {
        setError('加载日志数据失败')
        console.error('加载日志失败:', err)
        setLoading(false)
        // 清空数据避免使用旧数据
        setAllLogs([])
        setFilteredLogs([])
        setDisplayLogs([])
      }
    }
    
    fetchLogs()
  }, [selectedSchema, selectedModule, schemas, modules])

  // 更新显示的日志
  const updateDisplayLogs = (logs: LogEntry[], page: number, pageSize: number) => {
    const start = (page - 1) * pageSize
    const end = start + pageSize
    setDisplayLogs(logs.slice(start, end))
  }

  // 处理页面变化
  const handlePageChange = (page: number, pageSize: number = DEFAULT_PAGE_SIZE) => {
    const newPagination = {
      ...pagination,
      current: page,
      pageSize
    }
    setPagination(newPagination)
    updateDisplayLogs(filteredLogs, page, pageSize)
  }

  // 处理过滤
  const handleFilter = (filters: LogFilterOptions) => {
    setLoading(true)
    setCurrentFilters(filters)
    
    setTimeout(() => {
      try {
        // 应用过滤逻辑
        const filtered = allLogs.filter(log => {
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
            const logTime = new Date(log.timestamp).getTime()
            const startTime = new Date(filters.timeRange[0]).getTime()
            const endTime = new Date(filters.timeRange[1]).getTime()
            
            if (logTime < startTime || logTime > endTime) {
              return false
            }
          }
          
          // 搜索文本过滤
          if (filters.searchText && !log.message.toLowerCase().includes(filters.searchText.toLowerCase())) {
            return false
          }
          
          return true
        })
        
        setFilteredLogs(filtered)
        
        const newPagination = {
          ...pagination,
          current: 1,
          total: filtered.length
        }
        
        setPagination(newPagination)
        updateDisplayLogs(filtered, 1, pagination.pageSize)
        setLoading(false)
      } catch (err) {
        setError('过滤日志时出错')
        setLoading(false)
      }
    }, 300)
  }

  // 刷新日志数据
  const handleRefresh = () => {
    if (selectedSchema) {
      // 触发useEffect重新加载数据
      const schemaId = selectedSchema
      setSelectedSchema('')
      setTimeout(() => setSelectedSchema(schemaId), 0)
    }
  }

  return (
    <div className="site-layout-content">
      <div className="logs-page-header">
        <Title level={3}>日志管理</Title>
        
        {error && (
          <Alert
            message="错误"
            description={error}
            type="error"
            showIcon
            closable
            onClose={() => setError(null)}
            style={{ marginBottom: 16 }}
          />
        )}
        
        <div style={{ marginBottom: 16 }}>
          <Form layout="inline">
            <Form.Item label="Schema">
              <Select
                style={{ width: 200 }}
                value={selectedSchema}
                onChange={setSelectedSchema}
                placeholder="选择Schema"
              >
                {schemas.map(schema => (
                  <Option key={schema.id} value={schema.id}>{schema.name}</Option>
                ))}
              </Select>
            </Form.Item>
            
            <Form.Item label="模块">
              <Select
                style={{ width: 200 }}
                value={selectedModule}
                onChange={setSelectedModule}
                placeholder="选择模块"
                disabled={!selectedSchema || modules.length === 0}
                allowClear
              >
                {modules.map(module => (
                  <Option key={module.id} value={module.name}>{module.name}</Option>
                ))}
              </Select>
            </Form.Item>
            
            <Form.Item>
              <Button 
                type="primary" 
                icon={<ReloadOutlined />} 
                onClick={handleRefresh}
                loading={loading}
              >
                刷新数据
              </Button>
            </Form.Item>
          </Form>
        </div>
      </div>
      
      <div className="logs-filter-container">
        <Card title="日志过滤" variant="borderless">
          <LogFilter 
            services={services} 
            sources={sources}
            onFilter={handleFilter}
            loading={loading}
          />
        </Card>
      </div>
      
      <div className="logs-table-container">
        <Card 
          title={`日志列表 (${pagination.total}条记录)`} 
          variant="borderless"
          extra={loading && <Spin size="small" />}
          className="logs-table-card"
        >
          <div className="logs-table-scrollable">
            <LogTable 
              logs={displayLogs}
              pagination={false}
              onPageChange={handlePageChange}
              loading={loading}
              className="custom-table"
              scroll={{ y: '100%' }}
            />
          </div>
          <div className="logs-table-pagination">
            <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
              <Pagination
                current={pagination.current}
                pageSize={pagination.pageSize}
                total={pagination.total}
                onChange={handlePageChange}
                showSizeChanger={true}
                showTotal={(total: number) => `共 ${total} 条记录`}
              />
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}

export default LogsPage 