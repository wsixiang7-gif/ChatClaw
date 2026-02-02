<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Trash2, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { cn } from '@/lib/utils'
import { toast } from '@/components/ui/toast'
import { AgentsService, type Agent } from '@bindings/willchat/internal/services/agents'

type TabKey = 'model' | 'prompt' | 'delete'

const props = defineProps<{
  open: boolean
  agent: Agent | null
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  updated: [agent: Agent]
  deleted: [id: number]
}>()

const { t } = useI18n()

const tab = ref<TabKey>('model')
const saving = ref(false)
const deleteConfirmOpen = ref(false)

// prompt tab fields
const name = ref('')
const prompt = ref('')

// model tab fields
const temperatureEnabled = ref(true)
const topPEnabled = ref(true)
const temperature = ref(0.5)
const topP = ref(1.0)
const contextCount = ref(50)
const maxTokens = ref(1000)

watch(
  () => props.open,
  (open) => {
    if (!open) return
    tab.value = 'model'
  }
)

watch(
  () => props.agent,
  (agent) => {
    if (!agent) return
    name.value = agent.name ?? ''
    prompt.value = agent.prompt ?? ''

    temperature.value = agent.llm_temperature ?? 0.5
    topP.value = agent.llm_top_p ?? 1.0
    contextCount.value = agent.context_count ?? 50
    maxTokens.value = agent.llm_max_tokens ?? 1000
  },
  { immediate: true }
)

const isValid = computed(() => name.value.trim() !== '' && prompt.value.trim() !== '')

const handleClose = () => emit('update:open', false)

const handleSave = async () => {
  if (!props.agent || !isValid.value || saving.value) return
  saving.value = true
  try {
    const updated = await AgentsService.UpdateAgent(props.agent.id, {
      name: name.value.trim(),
      prompt: prompt.value.trim(),
      icon: null,
      default_llm_provider_id: null,
      default_llm_model_id: null,
      llm_temperature: temperatureEnabled.value ? temperature.value : null,
      llm_top_p: topPEnabled.value ? topP.value : null,
      context_count: contextCount.value,
      llm_max_tokens: maxTokens.value,
    })
    if (!updated) {
      throw new Error(t('assistant.errors.updateFailed'))
    }
    emit('updated', updated)
    toast.success(t('assistant.toasts.updated'))
    emit('update:open', false)
  } catch (e: any) {
    toast.error(e?.message ?? t('assistant.errors.updateFailed'))
  } finally {
    saving.value = false
  }
}

const handleDelete = async () => {
  if (!props.agent) return
  saving.value = true
  try {
    await AgentsService.DeleteAgent(props.agent.id)
    emit('deleted', props.agent.id)
    toast.success(t('assistant.toasts.deleted'))
    emit('update:open', false)
  } catch (e: any) {
    toast.error(e?.message ?? t('assistant.errors.deleteFailed'))
  } finally {
    saving.value = false
    deleteConfirmOpen.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="handleClose">
    <DialogContent class="sm:max-w-[820px]">
      <DialogHeader class="flex flex-row items-center justify-between">
        <DialogTitle>{{ t('assistant.settings.title') }}</DialogTitle>
        <Button size="icon" variant="ghost" @click="handleClose">
          <X class="size-4" />
        </Button>
      </DialogHeader>

      <div class="flex gap-6 py-2">
        <!-- 左侧 tabs -->
        <div class="flex w-[160px] shrink-0 flex-col gap-2">
          <button
            :class="
              cn(
                'rounded-md px-3 py-2 text-left text-sm transition-colors',
                tab === 'model'
                  ? 'bg-muted text-foreground'
                  : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
              )
            "
            @click="tab = 'model'"
          >
            {{ t('assistant.settings.tabs.model') }}
          </button>
          <button
            :class="
              cn(
                'rounded-md px-3 py-2 text-left text-sm transition-colors',
                tab === 'prompt'
                  ? 'bg-muted text-foreground'
                  : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
              )
            "
            @click="tab = 'prompt'"
          >
            {{ t('assistant.settings.tabs.prompt') }}
          </button>
          <button
            :class="
              cn(
                'rounded-md px-3 py-2 text-left text-sm transition-colors',
                tab === 'delete'
                  ? 'bg-muted text-foreground'
                  : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
              )
            "
            @click="tab = 'delete'"
          >
            {{ t('assistant.settings.tabs.delete') }}
          </button>
        </div>

        <!-- 右侧内容 -->
        <div class="min-w-0 flex-1">
          <!-- 模型设置 -->
          <div v-if="tab === 'model'" class="flex flex-col gap-5">
            <div class="flex flex-col gap-2">
              <div class="text-sm font-medium text-foreground">
                {{ t('assistant.settings.model.defaultModel') }}
              </div>
              <div class="flex items-center justify-between rounded-md border border-border px-3 py-2">
                <div class="text-sm text-foreground">
                  {{ props.agent?.default_llm_provider_id }}/{{ props.agent?.default_llm_model_id }}
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ t('assistant.settings.model.defaultModelHint') }}
                </div>
              </div>
            </div>

            <div class="flex items-center justify-between gap-4">
              <div class="flex flex-col">
                <div class="text-sm font-medium text-foreground">
                  {{ t('assistant.settings.model.temperature') }}
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ t('assistant.settings.model.temperatureHint') }}
                </div>
              </div>
              <Switch v-model:checked="temperatureEnabled" />
            </div>
            <div v-if="temperatureEnabled" class="flex items-center gap-3">
              <input
                v-model.number="temperature"
                type="range"
                min="0"
                max="2"
                step="0.05"
                class="w-full"
              />
              <div class="w-[60px] text-right text-sm text-muted-foreground">
                {{ temperature.toFixed(2) }}
              </div>
            </div>

            <div class="flex items-center justify-between gap-4">
              <div class="flex flex-col">
                <div class="text-sm font-medium text-foreground">
                  {{ t('assistant.settings.model.topP') }}
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ t('assistant.settings.model.topPHint') }}
                </div>
              </div>
              <Switch v-model:checked="topPEnabled" />
            </div>
            <div v-if="topPEnabled" class="flex items-center gap-3">
              <input v-model.number="topP" type="range" min="0" max="1" step="0.01" class="w-full" />
              <div class="w-[60px] text-right text-sm text-muted-foreground">
                {{ topP.toFixed(2) }}
              </div>
            </div>

            <div class="flex flex-col gap-2">
              <div class="text-sm font-medium text-foreground">
                {{ t('assistant.settings.model.contextCount') }}
              </div>
              <div class="flex items-center gap-3">
                <input v-model.number="contextCount" type="range" min="0" max="200" step="1" class="w-full" />
                <div class="w-[60px] text-right text-sm text-muted-foreground">
                  {{ contextCount }}
                </div>
              </div>
            </div>

            <div class="flex flex-col gap-1.5">
              <label class="text-sm font-medium text-foreground">
                {{ t('assistant.settings.model.maxTokens') }}
              </label>
              <Input v-model.number="maxTokens" type="number" min="1" max="200000" />
            </div>
          </div>

          <!-- 提示词设置 -->
          <div v-else-if="tab === 'prompt'" class="flex flex-col gap-4">
            <div class="flex flex-col gap-1.5">
              <label class="text-sm font-medium text-foreground">
                {{ t('assistant.fields.name') }}
                <span class="text-destructive">*</span>
              </label>
              <Input v-model="name" :placeholder="t('assistant.fields.namePlaceholder')" maxlength="100" />
            </div>

            <div class="flex flex-col gap-1.5">
              <label class="text-sm font-medium text-foreground">
                {{ t('assistant.fields.prompt') }}
                <span class="text-destructive">*</span>
              </label>
              <textarea
                v-model="prompt"
                :placeholder="t('assistant.fields.promptPlaceholder')"
                maxlength="1000"
                class="min-h-[260px] w-full resize-none rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              />
            </div>
          </div>

          <!-- 删除助手 -->
          <div v-else class="flex h-[320px] flex-col items-center justify-center gap-4">
            <div class="text-base font-semibold text-foreground">
              {{ t('assistant.settings.delete.title') }}
            </div>
            <div class="max-w-[420px] text-center text-sm text-muted-foreground">
              {{ t('assistant.settings.delete.hint') }}
            </div>

            <Button variant="destructive" :disabled="saving" @click="deleteConfirmOpen = true">
              <Trash2 class="mr-2 size-4" />
              {{ t('assistant.settings.delete.action') }}
            </Button>

            <AlertDialog :open="deleteConfirmOpen" @update:open="(v) => (deleteConfirmOpen = v)">
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>{{ t('assistant.settings.delete.confirmTitle') }}</AlertDialogTitle>
                  <AlertDialogDescription>
                    {{ t('assistant.settings.delete.confirmDesc', { name: props.agent?.name ?? '' }) }}
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel :disabled="saving">
                    {{ t('assistant.actions.cancel') }}
                  </AlertDialogCancel>
                  <AlertDialogAction class="bg-destructive text-destructive-foreground" :disabled="saving" @click="handleDelete">
                    {{ t('assistant.settings.delete.action') }}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" :disabled="saving" @click="handleClose">
          {{ t('assistant.actions.cancel') }}
        </Button>
        <Button v-if="tab !== 'delete'" :disabled="!isValid || saving" @click="handleSave">
          {{ t('assistant.actions.save') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

