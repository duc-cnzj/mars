import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: 'Mars',
  description: 'Built for DevOps. Ship a production-grade app in 30 seconds.',
  lang: 'en-US',
  base: '/mars/',

  // 国际化：默认语言（English）在站点根，中文镜像在 /zh/
  locales: {
    root: {
      lang: 'en-US',
      label: 'English',
      description: 'Built for DevOps. Ship a production-grade app in 30 seconds.',
      themeConfig: {
        nav: [
          { text: 'Home', link: '/' },
          { text: 'Docs', link: '/intro' }
        ],
        sidebar: [
          {
            text: 'Overview',
            items: [{ text: 'Introduction', link: '/intro' }]
          },
          {
            text: 'Quickstart',
            items: [
              { text: 'Quickstart', link: '/quickstart' },
              { text: 'Configuration', link: '/configuration' }
            ]
          },
          {
            text: 'Guides',
            items: [
              { text: 'Deploying Apps', link: '/projects' },
              { text: 'Containers & Logs', link: '/containers' },
              { text: 'Permissions', link: '/access-control' },
              { text: 'Audit & History', link: '/audit' }
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
    },
    zh: {
      lang: 'zh-CN',
      label: '简体中文',
      description: '专为 devops 而生，30 秒内部署一个应用。',
      themeConfig: {
        nav: [
          { text: '首页', link: '/zh/' },
          { text: '文档', link: '/zh/intro' }
        ],
        sidebar: [
          {
            text: '概览',
            items: [{ text: '简介', link: '/zh/intro' }]
          },
          {
            text: '快速开始',
            items: [
              { text: '快速开始', link: '/zh/quickstart' },
              { text: '系统配置', link: '/zh/configuration' }
            ]
          },
          {
            text: '使用指南',
            items: [
              { text: '部署应用', link: '/zh/projects' },
              { text: '容器与日志', link: '/zh/containers' },
              { text: '权限管理', link: '/zh/access-control' },
              { text: '审计与历史', link: '/zh/audit' }
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
