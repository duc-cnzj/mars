import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: 'Mars',
  description: '专为 devops 而生，30 秒内部署一个应用。',
  lang: 'zh-CN',
  base: '/mars/',

  // 国际化：默认语言（简体中文）在站点根，英文镜像在 /en/
  // locale key 约定：默认 locale 用 'root'，子 locale 用路径段（en 对应 docs/en/）
  locales: {
    root: {
      lang: 'zh-CN',
      label: '简体中文',
      description: '专为 devops 而生，30 秒内部署一个应用。',
      themeConfig: {
        nav: [
          { text: '首页', link: '/' },
          { text: '文档', link: '/intro' },
          { text: 'GitHub', link: 'https://github.com/duc-cnzj/mars' }
        ],
        sidebar: [
          {
            text: '概览',
            items: [{ text: '简介', link: '/intro' }]
          },
          {
            text: '快速开始',
            items: [
              { text: '快速开始', link: '/quickstart' },
              { text: '配置参考', link: '/configuration' }
            ]
          },
          {
            text: '使用指南',
            items: [
              { text: '项目管理', link: '/projects' },
              { text: '容器终端与日志', link: '/containers' },
              { text: '权限模型', link: '/access-control' },
              { text: '审计与变更记录', link: '/audit' }
            ]
          },
          {
            text: '参考',
            items: [
              { text: 'SDK 接入', link: '/sdk' },
              { text: 'API 参考', link: '/api-reference' }
            ]
          }
        ],
        docFooter: {
          prev: '上一页',
          next: '下一页'
        },
        footer: {
          copyright: 'created by duc · MIT License'
        }
      }
    },
    en: {
      lang: 'en-US',
      label: 'English',
      description: 'Built for DevOps. Ship a production-grade app in 30 seconds.',
      themeConfig: {
        nav: [
          { text: 'Home', link: '/en/' },
          { text: 'Docs', link: '/en/intro' },
          { text: 'GitHub', link: 'https://github.com/duc-cnzj/mars' }
        ],
        sidebar: [
          {
            text: 'Overview',
            items: [{ text: 'Introduction', link: '/en/intro' }]
          },
          {
            text: 'Quickstart',
            items: [
              { text: 'Quickstart', link: '/en/quickstart' },
              { text: 'Configuration', link: '/en/configuration' }
            ]
          },
          {
            text: 'Guides',
            items: [
              { text: 'Projects', link: '/en/projects' },
              { text: 'Containers & Logs', link: '/en/containers' },
              { text: 'Access Control', link: '/en/access-control' },
              { text: 'Audit & Changelog', link: '/en/audit' }
            ]
          },
          {
            text: 'Reference',
            items: [
              { text: 'SDK', link: '/en/sdk' },
              { text: 'API Reference', link: '/en/api-reference' }
            ]
          }
        ],
        docFooter: {
          prev: 'Previous page',
          next: 'Next page'
        },
        footer: {
          copyright: 'created by duc · MIT License'
        }
      }
    }
  },

  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    logo: {
      dark: '/dark-logo.png',
      light: '/logo512.png'
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/duc-cnzj/mars' }
    ]
  }
})
