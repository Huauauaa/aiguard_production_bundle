<template>
  <div v-if="show" class="log-viewer">
    <div class="log-header">
      <h3>运行日志</h3>
      <span class="log-path">{{ logPath }}</span>
    </div>
    <pre ref="logViewerRef" class="log-content">{{ content }}</pre>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  show: boolean
  logPath: string
  content: string
}>()

const logViewerRef = ref<HTMLElement | null>(null)

defineExpose({ logViewerRef })
</script>

<style scoped>
.log-viewer {
  margin-top: 1rem;
  border: 1px solid #334155;
  border-radius: 0.5rem;
  overflow: hidden;
}

.log-header {
  background: #1e293b;
  padding: 0.75rem 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.log-header h3 {
  margin: 0;
  font-size: 0.9rem;
}

.log-path {
  font-size: 0.75rem;
  color: #94a3b8;
}

.log-content {
  background: #0f172a;
  color: #e2e8f0;
  padding: 1rem;
  margin: 0;
  max-height: 300px;
  overflow-y: auto;
  font-size: 0.8rem;
  line-height: 1.5;
}
</style>
