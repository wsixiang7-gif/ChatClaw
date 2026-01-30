<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AcceptableValue } from 'reka-ui'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import SettingsCard from './SettingsCard.vue'
import SettingsItem from './SettingsItem.vue'

// 导入吸附应用图标
import WechatIcon from '@/assets/icons/snap/wechat.svg'
import WecomIcon from '@/assets/icons/snap/wecom.svg'
import QQIcon from '@/assets/icons/snap/qq.svg'
import DingtalkIcon from '@/assets/icons/snap/dingtalk.svg'
import FeishuIcon from '@/assets/icons/snap/feishu.svg'
import DouyinIcon from '@/assets/icons/snap/douyin.svg'

// 后端绑定
import { SettingsService, Category } from '@bindings/willchat/internal/services/settings'

const { t } = useI18n()

// 设置状态
const showAiSendButton = ref(true)
const sendKeyStrategy = ref('enter')
const showAiEditButton = ref(true)

// 吸附应用状态
const snapWechat = ref(true)
const snapWecom = ref(true)
const snapQQ = ref(true)
const snapDingtalk = ref(true)
const snapFeishu = ref(true)
const snapDouyin = ref(true)

// 发送按键模式选项
const sendKeyOptions = [
  { value: 'enter', label: 'settings.snap.sendKeyOptions.enter' },
  { value: 'ctrl_enter', label: 'settings.snap.sendKeyOptions.ctrlEnter' },
]

// 当前发送按键模式显示文本
const currentSendKeyLabel = computed(() => {
  const option = sendKeyOptions.find((opt) => opt.value === sendKeyStrategy.value)
  return option ? t(option.label) : ''
})

// 吸附应用列表
const snapApps = computed(() => [
  {
    key: 'snap_wechat',
    label: t('settings.snap.apps.wechat'),
    icon: WechatIcon,
    value: snapWechat,
  },
  {
    key: 'snap_wecom',
    label: t('settings.snap.apps.wecom'),
    icon: WecomIcon,
    value: snapWecom,
  },
  { key: 'snap_qq', label: t('settings.snap.apps.qq'), icon: QQIcon, value: snapQQ },
  {
    key: 'snap_dingtalk',
    label: t('settings.snap.apps.dingtalk'),
    icon: DingtalkIcon,
    value: snapDingtalk,
  },
  {
    key: 'snap_feishu',
    label: t('settings.snap.apps.feishu'),
    icon: FeishuIcon,
    value: snapFeishu,
  },
  {
    key: 'snap_douyin',
    label: t('settings.snap.apps.douyin'),
    icon: DouyinIcon,
    value: snapDouyin,
  },
])

// 加载设置
const loadSettings = async () => {
  try {
    const settings = await SettingsService.List(Category.CategorySnap)
    settings.forEach((setting) => {
      switch (setting.key) {
        case 'show_ai_send_button':
          showAiSendButton.value = setting.value === 'true'
          break
        case 'send_key_strategy':
          sendKeyStrategy.value = setting.value
          break
        case 'show_ai_edit_button':
          showAiEditButton.value = setting.value === 'true'
          break
        case 'snap_wechat':
          snapWechat.value = setting.value === 'true'
          break
        case 'snap_wecom':
          snapWecom.value = setting.value === 'true'
          break
        case 'snap_qq':
          snapQQ.value = setting.value === 'true'
          break
        case 'snap_dingtalk':
          snapDingtalk.value = setting.value === 'true'
          break
        case 'snap_feishu':
          snapFeishu.value = setting.value === 'true'
          break
        case 'snap_douyin':
          snapDouyin.value = setting.value === 'true'
          break
      }
    })
  } catch (error) {
    console.error('Failed to load snap settings:', error)
  }
}

// 更新设置
const updateSetting = async (key: string, value: string) => {
  try {
    await SettingsService.SetValue(key, value)
  } catch (error) {
    console.error(`Failed to update setting ${key}:`, error)
  }
}

// 处理 AI 发送按钮开关变化
const handleAiSendButtonChange = (val: boolean) => {
  showAiSendButton.value = val
  void updateSetting('show_ai_send_button', String(val))
}

// 处理 AI 编辑按钮开关变化
const handleAiEditButtonChange = (val: boolean) => {
  showAiEditButton.value = val
  void updateSetting('show_ai_edit_button', String(val))
}

// 处理吸附应用开关变化
const handleSnapAppChange = (key: string, refValue: { value: boolean }, val: boolean) => {
  refValue.value = val
  void updateSetting(key, String(val))
}

// 处理发送按键模式变化
const handleSendKeyChange = (value: AcceptableValue) => {
  if (typeof value === 'string') {
    sendKeyStrategy.value = value
    void updateSetting('send_key_strategy', value)
  }
}

// 页面加载时获取设置
onMounted(() => {
  void loadSettings()
})
</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- 设置卡片 -->
    <SettingsCard :title="t('settings.snap.title')">
      <!-- AI回复显示发送到聊天按钮 -->
      <SettingsItem :label="t('settings.snap.showAiSendButton')">
        <Switch :model-value="showAiSendButton" @update:model-value="handleAiSendButtonChange" />
      </SettingsItem>

      <!-- 发送消息按键模式 -->
      <SettingsItem :label="t('settings.snap.sendKeyStrategy')">
        <Select :model-value="sendKeyStrategy" @update:model-value="handleSendKeyChange">
          <SelectTrigger class="w-54">
            <SelectValue>{{ currentSendKeyLabel }}</SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem
              v-for="option in sendKeyOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ t(option.label) }}
            </SelectItem>
          </SelectContent>
        </Select>
      </SettingsItem>

      <!-- AI回复显示编辑内容按钮 -->
      <SettingsItem :label="t('settings.snap.showAiEditButton')" :bordered="false">
        <Switch :model-value="showAiEditButton" @update:model-value="handleAiEditButtonChange" />
      </SettingsItem>
    </SettingsCard>

    <!-- 吸附应用卡片 -->
    <SettingsCard :title="t('settings.snap.appsTitle')">
      <SettingsItem
        v-for="(app, index) in snapApps"
        :key="app.key"
        :bordered="index !== snapApps.length - 1"
      >
        <template #default>
          <Switch
            :model-value="app.value.value"
            @update:model-value="(val: boolean) => handleSnapAppChange(app.key, app.value, val)"
          />
        </template>
        <template #label>
          <div class="flex items-center gap-2">
            <component :is="app.icon" class="size-5" />
            <span class="text-sm font-medium text-foreground">{{ app.label }}</span>
          </div>
        </template>
      </SettingsItem>
    </SettingsCard>
  </div>
</template>
