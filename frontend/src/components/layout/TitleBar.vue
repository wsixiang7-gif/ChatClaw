<script setup lang="ts">
/**
 * 标题栏组件
 * 包含侧边栏折叠按钮和多标签页
 * 注意：macOS/Windows窗口控制由Wails自动处理，无需手动实现
 */
import { computed } from 'vue'
import { System } from '@wailsio/runtime'
import { useI18n } from 'vue-i18n'
import { useNavigationStore } from '@/stores'
import { cn } from '@/lib/utils'
import IconSidebarToggle from '@/assets/icons/sidebar-toggle.svg'
import IconAddNewTab from '@/assets/icons/add-new-tab.svg'

const navigationStore = useNavigationStore()
const { t } = useI18n()

const isMac = computed(() => System.IsMac())

/**
 * 处理标签页点击
 */
const handleTabClick = (tabId: string) => {
  navigationStore.setActiveTab(tabId)
}

/**
 * 处理标签页关闭
 */
const handleTabClose = (event: Event, tabId: string) => {
  event.stopPropagation()
  navigationStore.closeTab(tabId)
}

/**
 * 切换侧边栏折叠状态
 */
const handleToggleSidebar = () => {
  navigationStore.toggleSidebar()
}

/**
 * 新建一个 AI助手 标签页（右侧 + 按钮）
 */
const handleAddAssistantTab = () => {
  navigationStore.navigateToModule('assistant', t)
}
</script>

<template>
  <div
    class="flex h-10 items-end gap-2 overflow-hidden bg-[#dee8fa] pr-2 pt-1"
    style="--wails-draggable: drag"
  >
    <!-- macOS窗口控制按钮占位区域（Wails会自动渲染在这个位置） -->
    <!-- macOS需要约70px，Windows不需要占位 -->
    <div
      :class="cn('flex shrink-0 items-center self-stretch', isMac ? 'w-[70px]' : 'w-2')"
    />

    <!-- 侧边栏展开/收起按钮 -->
    <button
      :class="
        cn(
          'flex shrink-0 items-center self-stretch rounded p-1 hover:bg-[#ccddf5]',
          // macOS 下红黄绿的视觉中心略靠下，这里做 1px 微调以对齐
          isMac && 'translate-y-px'
        )
      "
      style="--wails-draggable: no-drag"
      @click="handleToggleSidebar"
    >
      <img :src="IconSidebarToggle" alt="" class="size-4" />
    </button>

    <!-- 标签页列表 -->
    <div class="z-2 flex shrink-0 items-end gap-1 self-stretch">
      <button
        v-for="tab in navigationStore.tabs"
        :key="tab.id"
        :class="
          cn(
            'group relative flex items-center justify-between gap-2 rounded-t-xl px-4 py-1.5',
            'transition-colors duration-150',
            navigationStore.activeTabId === tab.id
              ? 'bg-background text-foreground'
              : 'bg-[#dee8fa] text-muted-foreground hover:bg-[#ccddf5]'
          )
        "
        style="--wails-draggable: no-drag"
        @click="handleTabClick(tab.id)"
      >
        <div class="flex items-center gap-2">
          <!-- 标签页图标 -->
          <div
            v-if="tab.icon"
            class="size-5 shrink-0 overflow-hidden rounded-md"
          >
            <img :src="tab.icon" alt="" class="size-full object-cover" />
          </div>
          <div
            v-else
            class="flex size-5 shrink-0 items-center justify-center rounded-md bg-muted"
          >
            <svg
              viewBox="0 0 16 16"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
              class="size-3"
            >
              <path
                d="M8 2a6 6 0 100 12A6 6 0 008 2z"
                stroke="currentColor"
                stroke-width="1.5"
              />
            </svg>
          </div>
          <!-- 标签页标题 -->
          <span class="max-w-[100px] truncate text-sm">{{ tab.title }}</span>
        </div>
        <!-- 关闭按钮 -->
        <div
          class="flex size-4 items-center justify-center rounded opacity-0 transition-opacity hover:bg-muted group-hover:opacity-100"
          @click="handleTabClose($event, tab.id)"
        >
          <svg
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            class="size-3"
          >
            <path
              d="M4 4l8 8M12 4l-8 8"
              stroke="currentColor"
              stroke-width="1.5"
              stroke-linecap="round"
            />
          </svg>
        </div>
        <!-- 选中标签页的底部装饰（左右圆角连接） -->
        <template v-if="navigationStore.activeTabId === tab.id">
          <div class="absolute bottom-0 left-[-12px] size-3">
            <svg viewBox="0 0 12 12" fill="none" class="size-full">
              <path d="M12 12H0V0c0 6.627 5.373 12 12 12z" fill="white" />
            </svg>
          </div>
          <div class="absolute bottom-0 right-[-12px] size-3">
            <svg viewBox="0 0 12 12" fill="none" class="size-full">
              <path d="M0 12h12V0C12 6.627 6.627 12 0 12z" fill="white" />
            </svg>
          </div>
        </template>
      </button>

      <!-- + 按钮应紧挨最后一个标签页 -->
      <button
        class="flex shrink-0 items-center self-stretch rounded p-1 hover:bg-[#ccddf5]"
        style="--wails-draggable: no-drag"
        @click="handleAddAssistantTab"
      >
        <img :src="IconAddNewTab" alt="" class="size-4" />
      </button>
    </div>
  </div>
</template>
