import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import './i18n' // i18n 初始化：需在渲染前加载
import 'prism-themes/themes/prism-material-dark.css' // 语法高亮配色（旧版 index.tsx 同款，保证 yaml 等渲染一致）
import App from './App'
import { installPreloadRecovery } from './lib/preloadRecovery'

// 必须先于 render：下面每个路由页都是 React.lazy，首次点击就可能触发动态 import。
// 兜底监听要赶在第一个 chunk 请求之前就位，否则错过的事件无法补听（详见 preloadRecovery）。
installPreloadRecovery()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
