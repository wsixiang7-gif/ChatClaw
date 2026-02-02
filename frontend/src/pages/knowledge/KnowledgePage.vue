<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, MoreHorizontal, Settings } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { toast } from '@/components/ui/toast'
import CreateLibraryDialog from './components/CreateLibraryDialog.vue'
import EmbeddingSettingsDialog from './components/EmbeddingSettingsDialog.vue'

import type { Library } from '@bindings/willchat/internal/services/library'
import { LibraryService } from '@bindings/willchat/internal/services/library'

type LibraryTab = 'personal' | 'team'

const { t } = useI18n()

const activeTab = ref<LibraryTab>('personal')
const createDialogOpen = ref(false)
const embeddingSettingsOpen = ref(false)

const libraries = ref<Library[]>([])
const loading = ref(false)
const selectedLibraryId = ref<number | null>(null)

const selectedLibrary = computed(() =>
  libraries.value.find((l) => l.id === selectedLibraryId.value) || null
)

const getErrorMessage = (error: unknown): string => {
  let msg = ''
  if (error instanceof Error) msg = error.message
  else if (typeof error === 'string') msg = error
  else if (typeof error === 'object' && error !== null && 'message' in error) {
    msg = String((error as { message: unknown }).message)
  } else msg = String(error)

  if (msg.startsWith('{')) {
    try {
      const parsed = JSON.parse(msg)
      if (parsed.message) return parsed.message
    } catch {
      // ignore
    }
  }
  return msg
}

const loadLibraries = async () => {
  loading.value = true
  try {
    const list = await LibraryService.ListLibraries()
    libraries.value = list || []
    if (selectedLibraryId.value == null && libraries.value.length > 0) {
      selectedLibraryId.value = libraries.value[0].id
    }
  } catch (error) {
    console.error('Failed to load libraries:', error)
    toast.error(getErrorMessage(error) || t('knowledge.loadFailed'))
  } finally {
    loading.value = false
  }
}

const handleCreateClick = () => {
  createDialogOpen.value = true
}

const handleEmbeddingSettingsClick = () => {
  embeddingSettingsOpen.value = true
}

const handleCreated = (lib: Library) => {
  // 立即插入列表（减少一次刷新等待），并选中
  libraries.value = [...libraries.value, lib].sort(
    (a, b) => (b.sort_order - a.sort_order) || (b.id - a.id)
  )
  selectedLibraryId.value = lib.id
  toast.success(t('knowledge.create.success'))
}

onMounted(() => {
  void loadLibraries()
})
</script>

<template>
  <div class="flex h-full w-full bg-background text-foreground">
    <!-- 左侧：知识库列表 -->
    <aside class="flex w-[260px] shrink-0 flex-col border-r border-border">
      <div class="flex items-center justify-between gap-2 px-3 py-3">
        <!-- tabs -->
        <div class="flex items-center rounded-lg border border-border bg-muted/30 p-0.5">
          <button
            type="button"
            :class="
              cn(
                'h-8 rounded-md px-3 text-sm transition-colors',
                activeTab === 'personal'
                  ? 'bg-background text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              )
            "
            @click="activeTab = 'personal'"
          >
            {{ t('knowledge.tabs.personal') }}
          </button>
          <button
            type="button"
            disabled
            :class="
              cn(
                'h-8 rounded-md px-3 text-sm text-muted-foreground opacity-50 cursor-not-allowed'
              )
            "
            :title="t('knowledge.tabs.teamDisabledTip')"
          >
            {{ t('knowledge.tabs.team') }}
          </button>
        </div>

        <div v-if="activeTab === 'personal'" class="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            class="h-8 w-8"
            :title="t('knowledge.create.title')"
            @click="handleCreateClick"
          >
            <Plus class="size-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class="h-8 w-8"
            :title="t('knowledge.embeddingSettings.title')"
            @click="handleEmbeddingSettingsClick"
          >
            <Settings class="size-4" />
          </Button>
        </div>
      </div>

      <div class="flex-1 overflow-auto px-2 pb-2">
        <div v-if="loading" class="px-2 py-6 text-sm text-muted-foreground">
          {{ t('knowledge.loading') }}
        </div>

        <div
          v-else-if="activeTab === 'personal' && libraries.length === 0"
          class="mx-2 mt-2 rounded-lg border border-border bg-card p-4 text-sm text-muted-foreground"
        >
          <div class="font-medium text-foreground">{{ t('knowledge.empty.title') }}</div>
          <div class="mt-1">{{ t('knowledge.empty.desc') }}</div>
          <Button class="mt-3" size="sm" @click="handleCreateClick">
            <Plus class="mr-1.5 size-4" />
            {{ t('knowledge.create.title') }}
          </Button>
        </div>

        <div v-else class="flex flex-col gap-1">
          <button
            v-for="lib in libraries"
            :key="lib.id"
            type="button"
            :class="
              cn(
                'group flex w-full items-center gap-2 rounded-lg px-2 py-2 text-left text-sm transition-colors',
                selectedLibraryId === lib.id
                  ? 'bg-accent text-accent-foreground'
                  : 'text-foreground hover:bg-accent/50'
              )
            "
            @click="selectedLibraryId = lib.id"
          >
            <span class="min-w-0 flex-1 truncate">{{ lib.name }}</span>
            <div
              role="button"
              tabindex="0"
              class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground opacity-0 transition-opacity hover:bg-background/60 hover:text-foreground group-hover:opacity-100"
              :title="t('knowledge.item.menu')"
              @click.stop
            >
              <MoreHorizontal class="size-4" />
            </div>
          </button>
        </div>
      </div>
    </aside>

    <!-- 右侧：内容区（暂时占位） -->
    <main class="flex flex-1 flex-col overflow-auto">
      <div class="flex h-full w-full items-center justify-center px-8">
        <div
          v-if="activeTab !== 'personal'"
          class="rounded-2xl border border-border bg-card p-8 text-muted-foreground shadow-sm dark:border-white/15 dark:shadow-none dark:ring-1 dark:ring-white/5"
        >
          {{ t('knowledge.teamNotReady') }}
        </div>
        <div
          v-else-if="!selectedLibrary"
          class="rounded-2xl border border-border bg-card p-8 text-muted-foreground shadow-sm dark:border-white/15 dark:shadow-none dark:ring-1 dark:ring-white/5"
        >
          {{ t('knowledge.selectOne') }}
        </div>
        <div
          v-else
          class="w-full max-w-[720px] rounded-2xl border border-border bg-card p-8 shadow-sm dark:border-white/15 dark:shadow-none dark:ring-1 dark:ring-white/5"
        >
          <div class="text-lg font-semibold text-foreground">{{ selectedLibrary.name }}</div>
          <div class="mt-2 grid grid-cols-2 gap-3 text-sm text-muted-foreground">
            <div class="flex flex-col gap-1">
              <div class="text-xs">{{ t('knowledge.detail.embedding') }}</div>
              <div class="text-foreground">
                {{ selectedLibrary.embedding_provider_id }} / {{ selectedLibrary.embedding_model_id }}
              </div>
            </div>
            <div class="flex flex-col gap-1">
              <div class="text-xs">{{ t('knowledge.detail.rerank') }}</div>
              <div class="text-foreground">
                {{ selectedLibrary.rerank_provider_id }} / {{ selectedLibrary.rerank_model_id }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <CreateLibraryDialog v-model:open="createDialogOpen" @created="handleCreated" />
    <EmbeddingSettingsDialog v-model:open="embeddingSettingsOpen" />
  </div>
</template>
