import { Layout, Menu, Dropdown, Space, Button, Avatar } from 'antd'
import { UserOutlined, DownOutlined, DashboardOutlined, FileTextOutlined, SettingOutlined, BellOutlined } from '@ant-design/icons'
import './App.css'
import LogsPage from './pages/LogsPage'

const { Header, Content, Footer } = Layout

function App() {
  const userMenuItems = [
    {
      key: 'profile',
      label: '个人资料',
    },
    {
      key: 'settings',
      label: '设置',
    },
    {
      key: 'logout',
      label: '退出登录',
    },
  ]

  const mainMenuItems = [
    {
      key: 'dashboard',
      icon: <DashboardOutlined />,
      label: '仪表盘',
    },
    {
      key: 'logs',
      icon: <FileTextOutlined />,
      label: '日志管理',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置',
    },
  ]

  return (
    <Layout className="layout">
      <Header style={{ display: 'flex', alignItems: 'center' }}>
        <div className="logo">Agera Logs</div>
        <Menu
          theme="dark"
          mode="horizontal"
          defaultSelectedKeys={['logs']}
          items={mainMenuItems}
          style={{ flex: 1, minWidth: 0 }}
        />
        <div className="header-right">
          <Space size="middle">
            <Button type="text" icon={<BellOutlined />} style={{ color: 'white' }} />
            <Dropdown menu={{ items: userMenuItems }}>
              <a onClick={e => e.preventDefault()}>
                <Space>
                  <Avatar icon={<UserOutlined />} />
                  <span style={{ color: 'white' }}>管理员</span>
                  <DownOutlined style={{ color: 'white', fontSize: '12px' }} />
                </Space>
              </a>
            </Dropdown>
          </Space>
        </div>
      </Header>
      
      <Content>
        <LogsPage />
      </Content>
      
      <Footer style={{ textAlign: 'center' }}>
        Agera Logs ©{new Date().getFullYear()} 由 Ant Design 提供 UI 支持
      </Footer>
    </Layout>
  )
}

export default App
