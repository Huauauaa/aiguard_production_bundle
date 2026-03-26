<template>
  <div class="history-list">
    <h3>历史记录</h3>
    <div v-if="history.length === 0" class="empty">暂无历史记录</div>
    <div v-else class="history-items">
      <div v-for="item in history" :key="item.taskId" class="history-item">
        <div class="history-header">
          <span class="task-id">{{ item.taskId.slice(0, 8) }}</span>
          <span class="timestamp">{{ item.timestamp }}</span>
        </div>
        <div class="history-details">
          <span>{{ item.sourceBranch }} → {{ item.targetBranch }}</span>
          <span class="finding-count">{{ item.findingCount }} 个问题</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { HistoryItem } from '../types'

defineProps<{
  history: HistoryItem[]
}>()
</script>

<style scoped>
.history-list {
  margin-top: 1.5rem;
}

.history-list h3 {
  margin-bottom: 0.75rem;
  font-size: 1rem;
}

.empty {
  color: #64748b;
  font-size: 0.9rem;
  padding: 1rem;
  text-align: center;
  background: #1e293b;
  border-radius: 0.5rem;
}

.history-items {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.history-item {
  background: #1e293b;
  padding: 0.75rem;
  border-radius: 0.5rem;
  font-size: 0.85rem;
}

.history-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.task-id {
  font-family: monospace;
  color: #60a5fa;
}

.timestamp {
  color: #94a3b8;
  font-size: 0.8rem;
}

.history-details {
  display: flex;
  justify-content: space-between;
  color: #cbd5e1;
}

.finding-count {
  color: #fbbf24;
  font-weight: 600;
}
</style>
