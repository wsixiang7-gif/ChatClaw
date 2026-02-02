<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

const props = defineProps<{
  open: boolean
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  create: [data: { name: string; prompt: string }]
}>()

const { t } = useI18n()

const name = ref('')
const prompt = ref('')

const isValid = computed(() => name.value.trim() !== '' && prompt.value.trim() !== '')

watch(
  () => props.open,
  (open) => {
    if (!open) return
    name.value = ''
    prompt.value = ''
  }
)

const handleClose = () => emit('update:open', false)
const handleCreate = () => {
  if (!isValid.value || props.loading) return
  emit('create', { name: name.value.trim(), prompt: prompt.value.trim() })
}
</script>

<template>
  <Dialog :open="open" @update:open="handleClose">
    <DialogContent class="sm:max-w-[560px]">
      <DialogHeader>
        <DialogTitle>{{ t('assistant.create.title') }}</DialogTitle>
      </DialogHeader>

      <div class="flex flex-col gap-4 py-4">
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
            class="min-h-[160px] w-full resize-none rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" :disabled="loading" @click="handleClose">
          {{ t('assistant.actions.cancel') }}
        </Button>
        <Button :disabled="!isValid || loading" @click="handleCreate">
          {{ t('assistant.actions.create') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

