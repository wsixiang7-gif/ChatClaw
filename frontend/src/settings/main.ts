import { createApp } from "vue";
import App from "./App.vue";
import { i18n } from "../i18n";
import { initLocaleFromBackend } from "../composables/useLocale";

const app = createApp(App);
app.use(i18n);
app.mount("#app");

// 从后端同步语言设置
initLocaleFromBackend();

