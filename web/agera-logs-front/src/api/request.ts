import axios, { InternalAxiosRequestConfig, AxiosResponse, AxiosError, AxiosInstance } from 'axios'

// 扩展自定义Axios实例类型，确保返回类型正确
interface CustomAxiosInstance extends AxiosInstance {
  get<T = any>(url: string, config?: any): Promise<T>;
  post<T = any>(url: string, data?: any, config?: any): Promise<T>;
  put<T = any>(url: string, data?: any, config?: any): Promise<T>;
  delete<T = any>(url: string, config?: any): Promise<T>;
}

// 创建axios实例
const request = axios.create({
  baseURL: '/api', // 可以通过环境变量配置
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json'
  }
}) as CustomAxiosInstance

// 请求拦截器
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 可以在这里添加token等认证信息
    console.log(`准备发送请求: ${config.method?.toUpperCase()} ${config.url}`)
    return config
  },
  (error: AxiosError) => {
    return Promise.reject('请求配置错误: ' + error.message)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    console.log(`收到响应: ${response.status} ${response.config.url}`, response.data)
    
    if (response.data === null || response.data === undefined) {
      console.warn('响应数据为空，返回空数组')
      return []
    }
    
    // 如果是对象，但有error字段，表示后端返回了错误
    if (typeof response.data === 'object' && !Array.isArray(response.data) && response.data.error) {
      console.error('后端返回错误:', response.data.error)
      return Promise.reject(response.data.error)
    }
    
    return response.data
  },
  (error: AxiosError) => {
    let message: string

    if (error.code === 'ECONNABORTED') {
      message = '请求超时，请检查网络连接'
    } else if (error.message && error.message.includes('Network Error')) {
      message = '网络错误，无法连接到服务器，请确认后端服务是否启动'
    } else if (error.response) {
      // 服务器返回了错误状态码
      const status = error.response.status
      switch (status) {
        case 400:
          message = '请求参数有误'
          break
        case 401:
          message = '未授权，请重新登录'
          break
        case 403:
          message = '拒绝访问，没有权限'
          break
        case 404:
          message = '请求的资源不存在'
          break
        case 500:
          message = '服务器内部错误'
          break
        default:
          message = `服务器响应错误，状态码: ${status}`
      }

      // 如果服务器返回了错误信息，尝试提取
      if (error.response.data && typeof error.response.data === 'object') {
        const data = error.response.data as any
        if (data.error || data.message) {
          message += `: ${data.error || data.message}`
        }
      }
    } else if (error.request) {
      // 请求发出但未收到响应
      message = '服务器没有响应，请确认后端服务是否正常运行'
    } else {
      // 请求设置触发的错误
      message = '请求错误: ' + (error.message || '未知错误')
    }
    
    console.error('API请求失败:', message, error)
    
    // 返回空数组而不是拒绝Promise，这样调用方可以处理空结果而不是错误
    return Promise.resolve([])
  }
)

export default request 