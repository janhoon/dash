<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import Sidebar from './components/Sidebar.vue'
import { useAuth } from './composables/useAuth'

const route = useRoute()
const { isAuthenticated } = useAuth()

const sidebarRef = ref<InstanceType<typeof Sidebar> | null>(null)

const sidebarWidth = computed(() => {
  return sidebarRef.value?.isExpanded ? '200px' : '56px'
})

const showSidebar = computed(() => {
  return isAuthenticated.value && route.name !== 'login'
})
</script>

<template>
  <div class="app-layout" :class="{ 'no-sidebar': !showSidebar }">
    <Sidebar v-if="showSidebar" ref="sidebarRef" />
    <main class="main-content" :style="showSidebar ? { marginLeft: sidebarWidth } : {}">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  width: 100%;
}

.app-layout.no-sidebar {
  display: block;
}

.main-content {
  flex: 1;
  margin-left: 200px;
  min-height: 100vh;
  background: var(--bg-primary);
  transition: margin-left 0.2s ease;
}

.no-sidebar .main-content {
  margin-left: 0;
}
</style>
