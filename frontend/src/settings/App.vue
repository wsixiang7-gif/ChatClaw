<script setup lang="ts">
import { onMounted, ref } from "vue";
import { WindowService } from "../wails";

type WindowInfo = {
  name: string;
  title: string;
  url: string;
  created: boolean;
  visible: boolean;
};

const windows = ref<WindowInfo[]>([]);

const refresh = async () => {
  windows.value = await WindowService.List();
};

const hideSelf = async () => {
  await WindowService.Hide("settings");
};

onMounted(() => {
  void refresh();
});
</script>

<template>
  <div style="padding: 16px">
    <h2>设置</h2>

    <div class="card" style="margin: 12px 0">
      <div style="display: flex; gap: 8px; flex-wrap: wrap">
        <button class="btn" @click="refresh">刷新窗口列表</button>
        <button class="btn" @click="hideSelf">隐藏设置窗口</button>
      </div>
    </div>

    <div class="card">
      <div style="font-weight: 600; margin-bottom: 8px">窗口状态</div>
      <pre style="white-space: pre-wrap; margin: 0">{{ windows }}</pre>
    </div>
  </div>
</template>

