import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Mars",
  description: "专为devops而生，30秒内部署一个应用。",
  base: "/mars/",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    logo: '/logo512.png',

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
        text: '快速开始',
        items: [
          { text: '安装', link: '/install' },
          { text: '配置', link: '/configure' },
          { text: '让项目跑起来', link: '/run' },
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
