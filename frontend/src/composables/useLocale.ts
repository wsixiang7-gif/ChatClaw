import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { type Locale } from '../locales'
import { i18n } from '../i18n'
import { I18nService } from '../wails'

/**
 * 组件内使用的 locale composable
 */
export function useLocale() {
  const { locale } = useI18n()

  // 当前语言（响应式）
  const currentLocale = computed(() => locale.value as Locale)

  // 设置语言
  function setLocale(newLocale: Locale) {
    locale.value = newLocale
  }

  // 从后端同步语言设置
  async function syncFromBackend() {
    try {
      const backendLocale = await I18nService.GetLocale()
      if (backendLocale === 'zh-CN' || backendLocale === 'en-US') {
        setLocale(backendLocale as Locale)
      }
    } catch (e) {
      console.warn('Failed to sync locale from backend:', e)
    }
  }

  return {
    locale: currentLocale,
    setLocale,
    syncFromBackend,
  }
}

/**
 * 应用初始化时从后端同步语言（在 main.ts 中调用）
 */
export async function initLocaleFromBackend() {
  try {
    const backendLocale = await I18nService.GetLocale()
    if (backendLocale === 'zh-CN' || backendLocale === 'en-US') {
      i18n.global.locale.value = backendLocale
    }
  } catch (e) {
    console.warn('Failed to init locale from backend:', e)
  }
}
