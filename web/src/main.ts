import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './style.css'
import App from './App.vue'
import router from './router'

// Markdown Editor & Preview Setup (@kangc/v-md-editor)
import VMdEditor from '@kangc/v-md-editor'
import '@kangc/v-md-editor/lib/style/base-editor.css'
import VMdPreview from '@kangc/v-md-editor/lib/preview'
import '@kangc/v-md-editor/lib/style/preview.css'
import vuepressTheme from '@kangc/v-md-editor/lib/theme/vuepress.js'
import '@kangc/v-md-editor/lib/theme/style/vuepress.css'

import Prism from 'prismjs'

// Register vuepress theme for both editor and preview
VMdEditor.use(vuepressTheme, {
  Prism,
})

VMdPreview.use(vuepressTheme, {
  Prism,
})

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(VMdEditor)
app.use(VMdPreview)

app.mount('#app')
