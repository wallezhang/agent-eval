import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'AgentEval',
  description: 'A YAML-config-driven CLI tool for evaluating AI agents',

  locales: {
    en: {
      label: 'English',
      lang: 'en',
      link: '/en/',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/en/guide/quick-start' },
          { text: 'Reference', link: '/en/reference/configuration' },
          {
            text: 'GitHub',
            link: 'https://github.com/wallezhang/agent-eval'
          }
        ],
        sidebar: {
          '/en/guide/': [
            {
              text: 'Guide',
              items: [
                { text: 'Quick Start', link: '/en/guide/quick-start' },
                { text: 'Core Concepts', link: '/en/guide/concepts' },
                { text: 'Examples', link: '/en/guide/examples' },
                { text: 'Advanced Usage', link: '/en/guide/advanced' }
              ]
            }
          ],
          '/en/reference/': [
            {
              text: 'Reference',
              items: [
                { text: 'Configuration', link: '/en/reference/configuration' },
                { text: 'CLI', link: '/en/reference/cli' }
              ]
            }
          ]
        }
      }
    },
    zh: {
      label: '简体中文',
      lang: 'zh-CN',
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: '指南', link: '/zh/guide/quick-start' },
          { text: '参考', link: '/zh/reference/configuration' },
          {
            text: 'GitHub',
            link: 'https://github.com/wallezhang/agent-eval'
          }
        ],
        sidebar: {
          '/zh/guide/': [
            {
              text: '指南',
              items: [
                { text: '快速开始', link: '/zh/guide/quick-start' },
                { text: '核心概念', link: '/zh/guide/concepts' },
                { text: '示例', link: '/zh/guide/examples' },
                { text: '高级用法', link: '/zh/guide/advanced' }
              ]
            }
          ],
          '/zh/reference/': [
            {
              text: '参考',
              items: [
                { text: '配置参考', link: '/zh/reference/configuration' },
                { text: 'CLI 参考', link: '/zh/reference/cli' }
              ]
            }
          ]
        }
      }
    }
  },

  themeConfig: {
    socialLinks: [
      { icon: 'github', link: 'https://github.com/wallezhang/agent-eval' }
    ],
    search: {
      provider: 'local'
    }
  }
})
