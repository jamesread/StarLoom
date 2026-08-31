import { createApp } from 'vue'
import { initCustomTheme } from 'picocrank/vue/composables/useCustomTheme.js'
import './style.css'
import App from './App.vue'
import router from './router'
import { syncThemeFontStylesheet } from './lib/themeFonts'

initCustomTheme({
  storageKey: 'starapp-custom-theme',
  includeSupplementalThemes: true,
})

async function bootstrap() {
  if (typeof localStorage !== 'undefined') {
    const stored = (localStorage.getItem('starapp-custom-theme') || '').trim()
    if (stored) {
      await syncThemeFontStylesheet(stored)
    }
  }

  createApp(App).use(router).mount('#app')
}

void bootstrap()
