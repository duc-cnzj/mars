import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Mars",
  description: "专为devops而生，30秒内部署一个应用。",
  base: "/mars/",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    logo: {
      dark: '/dark-logo.png',
      light: '/logo512.png'
    },
    nav: [
      { text: 'Home', link: '/' },
      { text: '文档', link: '/intro' }
    ],

    sidebar: [
      {
        text: '概览',
        items: [
          { text: '简介', link: '/intro' },
        ]
      },
      {
        text: '入门',
        items: [
          { text: '快速体验', link: '/quick-start' },
          { text: '让项目跑起来', link: '/run' },
        ]
      },
      {
        text: '配置',
        items: [
          { text: '项目全局配置', link: '/configure' },
          { text: '单独分支配置', link: '/yaml-configure' },
          { text: 'annotations', link: '/annotations' },
          { text: '环境变量', link: '/env' },
        ]
      },
      {
        text: '可用插件',
        items: [
          { text: 'git 仓库', link: '/gitserver' },
          { text: '域名', link: '/domain' },
          { text: '登录页背景图', link: '/picture' },
          { text: 'Websocket', link: '/ws' },
        ]
      },
      {
        text: 'SDK 接入',
        items: [
          { text: 'Golang', link: '/go-sdk' },
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/duc-cnzj/mars' }
    ],
    footer: {
      copyright: 'created by duc@2023.'
    }
  }
})
