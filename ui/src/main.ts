import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

console.log(`
# AlbinoDrought/creamy-nvr
Repo: https://github.com/AlbinoDrought/creamy-nvr
Source: ${window.location.origin}/source.tar.gz
`);

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
