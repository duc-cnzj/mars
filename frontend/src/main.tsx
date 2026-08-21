import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import './i18n' // i18n 初始化：需在渲染前加载
import 'prism-themes/themes/prism-material-dark.css' // 语法高亮配色（旧版 index.tsx 同款，保证 yaml 等渲染一致）
import App from './App'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
