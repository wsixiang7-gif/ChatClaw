<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { cn } from '@/lib/utils'
import { useSettingsStore, type SettingsMenuItem } from '../stores/settings'

// 导入图标
import ModelServiceIcon from '@/assets/icons/model-service.svg'
import GeneralSettingsIcon from '@/assets/icons/general-settings.svg'
import SnapSettingsIcon from '@/assets/icons/snap-settings.svg'
import ToolsIcon from '@/assets/icons/tools.svg'
import AboutIcon from '@/assets/icons/about.svg'

const { t } = useI18n()
const settingsStore = useSettingsStore()

interface MenuItem {
  id: SettingsMenuItem
  labelKey: string
  icon: string
}

const menuItems: MenuItem[] = [
  { id: 'modelService', labelKey: 'settings.menu.modelService', icon: ModelServiceIcon },
  { id: 'generalSettings', labelKey: 'settings.menu.generalSettings', icon: GeneralSettingsIcon },
  { id: 'snapSettings', labelKey: 'settings.menu.snapSettings', icon: SnapSettingsIcon },
  { id: 'tools', labelKey: 'settings.menu.tools', icon: ToolsIcon },
  { id: 'about', labelKey: 'settings.menu.about', icon: AboutIcon },
]

const handleMenuClick = (menuId: SettingsMenuItem) => {
  settingsStore.setActiveMenu(menuId)
}
</script>

<template>
  <nav class="flex h-full w-[182px] flex-col gap-0.5 border-r border-border bg-background py-2">
    <button
      v-for="item in menuItems"
      :key="item.id"
      :class="
        cn(
          'mx-2 flex h-9 items-center gap-2.5 rounded-md px-2.5 text-left text-sm transition-colors',
          settingsStore.activeMenu === item.id
            ? 'bg-accent font-medium text-foreground'
            : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
        )
      "
      @click="handleMenuClick(item.id)"
    >
      <img
        :src="item.icon"
        :alt="t(item.labelKey)"
        class="size-4"
        :class="settingsStore.activeMenu === item.id ? 'opacity-100' : 'opacity-70'"
      />
      <span>{{ t(item.labelKey) }}</span>
    </button>
  </nav>
</template>
