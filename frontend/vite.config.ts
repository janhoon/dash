import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import monacoEditorPluginModule from 'vite-plugin-monaco-editor'

// Handle both ESM and CommonJS default export
const monacoEditorPlugin = (monacoEditorPluginModule as any).default || monacoEditorPluginModule

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    monacoEditorPlugin({
      languageWorkers: ['editorWorkerService'],
      customWorkers: []
    })
  ],
  server: {
    port: 5173
  },
  test: {
    environment: 'happy-dom',
    globals: true
  }
})
