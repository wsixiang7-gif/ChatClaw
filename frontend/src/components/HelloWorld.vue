<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {Events} from "@wailsio/runtime";
import {GreetService, WindowService} from "../wails";

defineProps<{ msg: string }>()

const name = ref('')
const result = ref('Please enter your name below 👇')
const time = ref('Listening for Time event...')
const settingsVisible = ref(false)

const doGreet = () => {
  let localName = name.value;
  if (!localName) {
    localName = 'anonymous';
  }
  GreetService.Greet(localName).then((resultValue: string) => {
    result.value = resultValue;
  }).catch((err: Error) => {
    console.log(err);
  });
}

onMounted(() => {
  Events.On('time', (timeValue: { data: string }) => {
    time.value = timeValue.data;
  });

  WindowService.IsVisible("settings").then((v: boolean) => {
    settingsVisible.value = v;
  }).catch(() => {
    settingsVisible.value = false;
  })
})

const showSettings = () => {
  WindowService.Show("settings").then(() => {
    settingsVisible.value = true;
  }).catch((err: Error) => console.log(err));
}

const hideSettings = () => {
  WindowService.Close("settings").then(() => {
    settingsVisible.value = false;
  }).catch((err: Error) => console.log(err));
}

</script>

<template>
  <h1>{{ msg }}</h1>

  <div aria-label="result" class="result">{{ result }}</div>
  <div class="card">
    <div class="input-box">
      <input aria-label="input" class="input" v-model="name" type="text" autocomplete="off"/>
      <button aria-label="greet-btn" class="btn" @click="doGreet">Greet</button>
    </div>
  </div>

  <div class="card" style="margin-top: 12px">
    <div style="display: flex; gap: 8px; flex-wrap: wrap">
      <button class="btn" @click="showSettings">Show Settings</button>
      <button class="btn" @click="hideSettings" :disabled="!settingsVisible">Hide Settings</button>
    </div>
  </div>

  <div class="footer">
    <div><p>Click on the Wails logo to learn more</p></div>
    <div><p>{{ time }}</p></div>
  </div>
</template>
