<script setup lang="ts">
/**
 * 设置页面组件
 * 可在主窗口和独立设置窗口中复用
 */
import { computed } from 'vue'
import SettingsSidebar from './components/SettingsSidebar.vue'
import GeneralSettings from './components/GeneralSettings.vue'
import { useSettingsStore } from './stores/settings'

const settingsStore = useSettingsStore()

// 根据当前菜单返回对应的内容组件
const currentComponent = computed(() => {
  switch (settingsStore.activeMenu) {
    case 'generalSettings':
      return GeneralSettings
    // 其他菜单页面后续实现
    default:
      return null
  }
})
</script>

<template>
  <div class="flex h-full w-full bg-background text-foreground">
    <!-- 侧边栏导航 -->
    <SettingsSidebar />

    <!-- 内容区域 -->
    <main class="flex flex-1 flex-col items-center overflow-auto py-8">
      <component :is="currentComponent" v-if="currentComponent" />
      <!-- 占位内容：当其他菜单页面还没实现时显示 -->
      <div
        v-else
        class="flex w-[530px] items-center justify-center rounded-2xl border border-border bg-card p-8 text-muted-foreground shadow-sm"
      >
        {{ settingsStore.activeMenu }}
      </div>
    </main>
  </div>
</template>
