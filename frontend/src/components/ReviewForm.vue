<template>
  <div class="review-form">
    <div class="field">
      <label>MR/PR 链接</label>
      <input v-model="form.mrUrl" placeholder="https://gitlab.example.com/group/project/-/merge_requests/123" />
    </div>

    <div class="field">
      <label>仓库地址（可选）</label>
      <input v-model="form.repoUrl" placeholder="git@gitlab.example.com:group/project.git" />
    </div>

    <div class="field">
      <label>本地仓库路径（可选）</label>
      <input v-model="form.localRepoPath" placeholder="/path/to/repo" />
    </div>

    <datalist id="branch-options">
      <option v-for="branch in availableBranches" :key="branch" :value="branch" />
    </datalist>

    <div class="field-row">
      <div class="field">
        <label>源分支</label>
        <input v-model="form.sourceBranch" list="branch-options" />
      </div>
      <div class="field">
        <label>目标分支</label>
        <input v-model="form.targetBranch" list="branch-options" />
      </div>
    </div>

    <div class="field-row">
      <div class="field">
        <label>配置文件路径</label>
        <input v-model="form.configPath" placeholder="./config.yaml" />
      </div>
      <div class="field">
        <label>工作区路径</label>
        <input v-model="form.workspaceDir" placeholder="./workspace" />
      </div>
    </div>

    <div class="action-grid">
      <button @click="$emit('pull-code')" :disabled="disabled">
        {{ pullingCode ? '拉取中...' : '拉取代码' }}
      </button>
      <button @click="$emit('start-review')" :disabled="disabled">
        {{ running ? '监视中...' : '开始监视' }}
      </button>
      <button class="secondary" @click="$emit('cancel-review')" :disabled="!running">取消任务</button>
      <button class="secondary" @click="$emit('toggle-logs')">{{ showLogs ? '隐藏日志' : '查看日志' }}</button>
      <button class="secondary danger" @click="$emit('clear-cache')" :disabled="disabled">
        {{ clearingCache ? '清理中...' : '清理缓存' }}
      </button>
      <button class="secondary" @click="$emit('load-history')">刷新历史</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { StartReviewRequest } from '../types'

defineProps<{
  form: StartReviewRequest
  availableBranches: string[]
  running: boolean
  pullingCode: boolean
  clearingCache: boolean
  showLogs: boolean
  disabled: boolean
}>()

defineEmits<{
  'pull-code': []
  'start-review': []
  'cancel-review': []
  'toggle-logs': []
  'clear-cache': []
  'load-history': []
}>()
</script>

<style scoped>
.review-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.field label {
  font-size: 0.9rem;
  font-weight: 500;
  color: #cbd5e1;
}

.field input {
  padding: 0.6rem;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 0.375rem;
  color: #e2e8f0;
  font-size: 0.9rem;
}

.field input:focus {
  outline: none;
  border-color: #3b82f6;
}

.field-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
  margin-top: 0.5rem;
}

button {
  padding: 0.6rem 1rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.375rem;
  font-size: 0.9rem;
  cursor: pointer;
  transition: background 0.2s;
}

button:hover:not(:disabled) {
  background: #2563eb;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

button.secondary {
  background: #475569;
}

button.secondary:hover:not(:disabled) {
  background: #334155;
}

button.danger {
  background: #dc2626;
}

button.danger:hover:not(:disabled) {
  background: #b91c1c;
}
</style>
