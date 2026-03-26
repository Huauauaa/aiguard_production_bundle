<template>
  <div v-if="progress" class="progress-panel">
    <div class="progress-header">
      <span class="stage">{{ progress.stage }}</span>
      <span class="percent">{{ progress.percent }}%</span>
    </div>
    <div class="progress-bar">
      <div class="progress-fill" :style="{ width: progress.percent + '%' }"></div>
    </div>
    <p class="progress-message">{{ progress.message }}</p>
  </div>
</template>

<script setup lang="ts">
import type { ProgressEvent } from '../types'

defineProps<{
  progress: ProgressEvent | null
}>()
</script>

<style scoped>
.progress-panel {
  background: #1e293b;
  padding: 1rem;
  border-radius: 0.5rem;
  margin-top: 1rem;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.stage {
  font-weight: 600;
  color: #60a5fa;
}

.percent {
  color: #94a3b8;
}

.progress-bar {
  height: 8px;
  background: #0f172a;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #60a5fa);
  transition: width 0.3s ease;
}

.progress-message {
  margin: 0.5rem 0 0;
  font-size: 0.85rem;
  color: #cbd5e1;
}
</style>
