<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { backend } from './bridge'
import type { HistoryItem, LogState, ProgressEvent, ReviewDoneEvent, StartReviewRequest } from './types'
import ReviewForm from './components/ReviewForm.vue'
import ProgressPanel from './components/ProgressPanel.vue'
import LogViewer from './components/LogViewer.vue'
import HistoryList from './components/HistoryList.vue'

const form = ref<StartReviewRequest>({
  mrUrl: '',
  repoUrl: '',
  localRepoPath: '',
  sourceBranch: '',
  targetBranch: '',
  configPath: 'config.yaml',
  workspaceDir: './workspace'
})

const currentTaskId = ref('')
const progress = ref<ProgressEvent | null>(null)
const doneEvent = ref<ReviewDoneEvent | null>(null)
const history = ref<HistoryItem[]>([])
const running = ref(false)
const pullingCode = ref(false)
const clearingCache = ref(false)
const errorMessage = ref('')
const infoMessage = ref('')
const availableBranches = ref<string[]>([])
const showLogs = ref(false)
const logState = ref<LogState>({ logPath: '', content: '', updatedAt: '' })
const logViewerRef = ref<InstanceType<typeof LogViewer> | null>(null)

let logTimer: number | null = null

const runtimeRequest = computed(() => ({
  configPath: form.value.configPath,
  workspaceDir: form.value.workspaceDir
}))

const findings = computed(() => doneEvent.value?.report.findings ?? [])
const report = computed(() => doneEvent.value?.report)
const disabled = computed(() => running.value || pullingCode.value || clearingCache.value)

function normalizeError(err: unknown): string {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return '操作失败，请查看日志了解详情。'
}

function clearMessages() {
  errorMessage.value = ''
  infoMessage.value = ''
}

function setInfo(message: string) {
  infoMessage.value = message
  errorMessage.value = ''
}

function setError(message: string) {
  errorMessage.value = message
}

async function loadHistory() {
  history.value = await backend.listHistory(runtimeRequest.value)
}

async function loadLogState(forceScroll = false) {
  logState.value = await backend.getLogState(runtimeRequest.value)
  if (showLogs.value || forceScroll) {
    await nextTick()
    if (logViewerRef.value?.logViewerRef) {
      logViewerRef.value.logViewerRef.scrollTop = logViewerRef.value.logViewerRef.scrollHeight
    }
  }
}

function stopLogPolling() {
  if (logTimer !== null) {
    window.clearInterval(logTimer)
    logTimer = null
  }
}

function startLogPolling() {
  stopLogPolling()
  logTimer = window.setInterval(() => void loadLogState(true), 1000)
}

async function pullCode() {
  clearMessages()
  pullingCode.value = true
  try {
    const result = await backend.pullCode(form.value)
    if (result.repoUrl) form.value.repoUrl = result.repoUrl
    form.value.sourceBranch = result.sourceBranch || form.value.sourceBranch
    form.value.targetBranch = result.targetBranch || form.value.targetBranch
    availableBranches.value = result.availableBranches ?? []
    setInfo(result.message || '代码拉取完成。')
    await Promise.all([loadHistory(), loadLogState(true)])
  } catch (err) {
    setError(normalizeError(err))
  } finally {
    pullingCode.value = false
  }
}

async function startReview() {
  clearMessages()
  doneEvent.value = null
  progress.value = null
  running.value = true
  try {
    currentTaskId.value = await backend.startReview(form.value)
    setInfo('监视已启动，正在执行前置校验与审计流程。')
    if (showLogs.value) await loadLogState(true)
  } catch (err) {
    running.value = false
    setError(normalizeError(err))
  }
}

async function cancelReview() {
  if (!currentTaskId.value) return
  clearMessages()
  try {
    await backend.cancelReview(currentTaskId.value)
    running.value = false
    currentTaskId.value = ''
    setInfo('任务已取消。')
    await loadLogState(true)
  } catch (err) {
    setError(normalizeError(err))
  }
}

async function clearCache() {
  clearMessages()
  clearingCache.value = true
  try {
    const result = await backend.clearCache(runtimeRequest.value)
    setInfo(result.message || '缓存已清理。')
    history.value = []
    await loadLogState(true)
  } catch (err) {
    setError(normalizeError(err))
  } finally {
    clearingCache.value = false
  }
}

function toggleLogs() {
  showLogs.value = !showLogs.value
  if (showLogs.value) {
    void loadLogState(true)
    startLogPolling()
  } else {
    stopLogPolling()
  }
}

onMounted(() => {
  void loadHistory()
  void loadLogState()
  window.runtime.EventsOn('review:progress', (event: ProgressEvent) => {
    progress.value = event
  })
  window.runtime.EventsOn('review:error', (event: { taskId: string; message: string }) => {
    running.value = false
    setError(event.message)
    stopLogPolling()
  })
  window.runtime.EventsOn('review:done', (event: ReviewDoneEvent) => {
    running.value = false
    doneEvent.value = event
    stopLogPolling()
    void loadHistory()
  })
})

onUnmounted(() => {
  stopLogPolling()
})
</script>

<template>
  <div class="app">
    <header class="header">
      <h1>AI代码监视</h1>
      <p class="subtitle">基于 LLM 的代码审计工具</p>
    </header>

    <main class="main">
      <div v-if="errorMessage" class="message error">{{ errorMessage }}</div>
      <div v-if="infoMessage" class="message info">{{ infoMessage }}</div>

      <ReviewForm
        :form="form"
        :available-branches="availableBranches"
        :running="running"
        :pulling-code="pullingCode"
        :clearing-cache="clearingCache"
        :show-logs="showLogs"
        :disabled="disabled"
        @pull-code="pullCode"
        @start-review="startReview"
        @cancel-review="cancelReview"
        @toggle-logs="toggleLogs"
        @clear-cache="clearCache"
        @load-history="loadHistory"
      />

      <ProgressPanel :progress="progress" />

      <LogViewer
        ref="logViewerRef"
        :show="showLogs"
        :log-path="logState.logPath"
        :content="logState.content"
      />

      <div v-if="report" class="report-section">
        <h2>审计报告</h2>
        <div class="report-summary">
          <div class="summary-item">
            <span class="label">总问题数</span>
            <span class="value">{{ findings.length }}</span>
          </div>
          <div class="summary-item">
            <span class="label">严重</span>
            <span class="value critical">{{ findings.filter(f => f.severity === '严重').length }}</span>
          </div>
          <div class="summary-item">
            <span class="label">高危</span>
            <span class="value high">{{ findings.filter(f => f.severity === '高危').length }}</span>
          </div>
        </div>

        <div v-if="findings.length > 0" class="findings-list">
          <div v-for="(finding, idx) in findings" :key="idx" class="finding-item">
            <div class="finding-header">
              <span :class="['severity', finding.severity]">{{ finding.severity }}</span>
              <span class="category">{{ finding.category }}</span>
            </div>
            <h3>{{ finding.title }}</h3>
            <p class="description">{{ finding.description }}</p>
            <div class="location">{{ finding.file }}:{{ finding.line }}</div>
          </div>
        </div>
      </div>

      <HistoryList :history="history" />
    </main>
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
  background: #0f172a;
  color: #e2e8f0;
}

.header {
  background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
  padding: 2rem;
  text-align: center;
  border-bottom: 2px solid #3b82f6;
}

.header h1 {
  margin: 0;
  font-size: 2rem;
  color: #60a5fa;
}

.subtitle {
  margin: 0.5rem 0 0;
  color: #94a3b8;
  font-size: 0.95rem;
}

.main {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.message {
  padding: 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.message.error {
  background: #7f1d1d;
  border: 1px solid #dc2626;
  color: #fecaca;
}

.message.info {
  background: #1e3a8a;
  border: 1px solid #3b82f6;
  color: #bfdbfe;
}

.report-section {
  margin-top: 2rem;
  background: #1e293b;
  padding: 1.5rem;
  border-radius: 0.5rem;
}

.report-section h2 {
  margin: 0 0 1rem;
  font-size: 1.25rem;
}

.report-summary {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.summary-item {
  background: #0f172a;
  padding: 1rem;
  border-radius: 0.375rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.summary-item .label {
  font-size: 0.85rem;
  color: #94a3b8;
}

.summary-item .value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #60a5fa;
}

.summary-item .value.critical {
  color: #ef4444;
}

.summary-item .value.high {
  color: #f59e0b;
}

.findings-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.finding-item {
  background: #0f172a;
  padding: 1rem;
  border-radius: 0.375rem;
  border-left: 3px solid #3b82f6;
}

.finding-header {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.severity {
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.severity.严重 {
  background: #7f1d1d;
  color: #fecaca;
}

.severity.高危 {
  background: #78350f;
  color: #fed7aa;
}

.category {
  padding: 0.25rem 0.5rem;
  background: #1e3a8a;
  color: #bfdbfe;
  border-radius: 0.25rem;
  font-size: 0.75rem;
}

.finding-item h3 {
  margin: 0 0 0.5rem;
  font-size: 1rem;
  color: #f1f5f9;
}

.description {
  margin: 0 0 0.5rem;
  font-size: 0.85rem;
  color: #cbd5e1;
  line-height: 1.5;
}

.location {
  font-family: monospace;
  font-size: 0.8rem;
  color: #64748b;
}
</style>
