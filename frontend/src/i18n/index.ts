import { createI18n } from 'vue-i18n'
import { messages } from '../locales'

// 创建 i18n 实例，默认使用中文
export const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: 'zh-CN',
  fallbackLocale: 'en-US',
  messages,
})
