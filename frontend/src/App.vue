<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { backend } from './bridge'
import type { HistoryItem, ProgressEvent, ReviewDoneEvent, StartReviewRequest } from './types'

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
const errorMessage = ref('')

const findings = computed(() => doneEvent.value?.report.findings ?? [])
const report = computed(() => doneEvent.value?.report)

async function startReview() {
  errorMessage.value = ''
  doneEvent.value = null
  progress.value = null
  running.value = true
  currentTaskId.value = await backend.startReview(form.value)
}

async function refreshSameTask() {
  await startReview()
}

async function cancelReview() {
  if (!currentTaskId.value) return
  await backend.cancelReview(currentTaskId.value)
  running.value = false
}

async function loadHistory() {
  history.value = await backend.listHistory()
}

onMounted(async () => {
  backend.on('review:progress', (payload: ProgressEvent) => {
    if (!currentTaskId.value || payload.taskId === currentTaskId.value) {
      progress.value = payload
    }
  })

  backend.on('review:done', async (payload: ReviewDoneEvent) => {
    if (!currentTaskId.value || payload.taskId === currentTaskId.value) {
      doneEvent.value = payload
      running.value = false
      await loadHistory()
    }
  })

  backend.on('review:error', (payload: { taskId: string; message: string }) => {
    if (!currentTaskId.value || payload.taskId === currentTaskId.value) {
      errorMessage.value = payload.message
      running.value = false
    }
  })

  await loadHistory()
})
</script>

<template>
  <div class="app-shell">
    <div class="hero">
      <h1>AI代码监视</h1>
      <p>面向 GitHub PR / GitLab MR / 本地仓库改动的桌面审计工具首版源码。</p>
    </div>

    <div class="layout">
      <div class="card">
        <div class="field">
          <label>MR / PR 链接</label>
          <input v-model="form.mrUrl" placeholder="https://github.com/org/repo/pull/123" />
        </div>
        <div class="field">
          <label>仓库地址（可选，链接无法自动识别时填写）</label>
          <input v-model="form.repoUrl" placeholder="https://github.com/org/repo.git" />
        </div>
        <div class="field">
          <label>本地仓库路径（可选，本地模式）</label>
          <input v-model="form.localRepoPath" placeholder="D:/code/project 或 /Users/me/code/project" />
        </div>
        <div class="field">
          <label>源分支</label>
          <input v-model="form.sourceBranch" placeholder="feature/login" />
        </div>
        <div class="field">
          <label>目标分支</label>
          <input v-model="form.targetBranch" placeholder="main" />
        </div>
        <div class="field">
          <label>配置文件路径</label>
          <input v-model="form.configPath" placeholder="./config.yaml" />
        </div>
        <div class="field">
          <label>工作区路径</label>
          <input v-model="form.workspaceDir" placeholder="./workspace" />
        </div>

        <div class="form-actions">
          <button @click="startReview" :disabled="running">开始监视</button>
          <button class="secondary" @click="refreshSameTask" :disabled="running">刷新代码</button>
          <button class="secondary" @click="cancelReview" :disabled="!running">取消任务</button>
          <button class="secondary" @click="loadHistory">刷新历史</button>
        </div>

        <div v-if="progress" class="status-line">
          <div class="small">阶段：{{ progress.stage }}</div>
          <div class="progress"><div :style="{ width: `${progress.percent}%` }" /></div>
          <div class="small">进度：{{ progress.percent }}% · {{ progress.message }}</div>
        </div>

        <div v-if="errorMessage" class="error">
          {{ errorMessage }}
        </div>

        <div style="margin-top: 20px">
          <h3>历史记录</h3>
          <ul class="history-list">
            <li v-for="item in history" :key="item.taskId">
              <strong>{{ item.title }}</strong><br />
              <span class="small">{{ item.createdAt }} · {{ item.sourceRef }} → {{ item.targetRef }} · 总问题 {{ item.totalIssues }}</span>
            </li>
            <li v-if="history.length === 0" class="placeholder">暂无历史记录</li>
          </ul>
        </div>
      </div>

      <div>
        <div class="card">
          <div class="report-header">
            <div>
              <h2 style="margin:0">审计概览</h2>
              <div class="small">审计过程中会持续推送阶段进度，结束后展示最新报告。</div>
            </div>
            <div v-if="doneEvent" class="small">
              报告目录：{{ doneEvent.reportDir }}
            </div>
          </div>

          <div class="metrics">
            <div class="metric">
              <div class="label">高（高危+严重）</div>
              <div class="value">{{ report?.summary.high ?? 0 }}</div>
            </div>
            <div class="metric">
              <div class="label">中（一般）</div>
              <div class="value">{{ report?.summary.medium ?? 0 }}</div>
            </div>
            <div class="metric">
              <div class="label">低（建议）</div>
              <div class="value">{{ report?.summary.low ?? 0 }}</div>
            </div>
            <div class="metric">
              <div class="label">总计</div>
              <div class="value">{{ report?.summary.total ?? 0 }}</div>
            </div>
          </div>

          <div v-if="report" style="margin-top: 18px" class="small">
            <div>HTML：{{ doneEvent?.htmlPath }}</div>
            <div>Markdown：{{ doneEvent?.markdownPath }}</div>
            <div>JSON：{{ doneEvent?.jsonPath }}</div>
          </div>
        </div>

        <div class="card" v-if="report">
          <h2 style="margin-top:0">质量与健康度</h2>
          <div class="metrics" style="grid-template-columns:repeat(5,minmax(0,1fr))">
            <div class="metric">
              <div class="label">安全性</div>
              <div class="value">{{ report.health.security }}</div>
            </div>
            <div class="metric">
              <div class="label">性能</div>
              <div class="value">{{ report.health.performance }}</div>
            </div>
            <div class="metric">
              <div class="label">健壮性</div>
              <div class="value">{{ report.health.robustness }}</div>
            </div>
            <div class="metric">
              <div class="label">可维护性</div>
              <div class="value">{{ report.health.maintainability }}</div>
            </div>
            <div class="metric">
              <div class="label">框架最佳实践</div>
              <div class="value">{{ report.health.frameworkPractice }}</div>
            </div>
          </div>

          <h3 style="margin-top:20px">其他说明</h3>
          <ul class="note-list">
            <li v-for="note in report.notes" :key="note">{{ note }}</li>
          </ul>
        </div>

        <div class="card">
          <h2 style="margin-top:0">问题清单</h2>
          <div v-if="findings.length === 0" class="placeholder">暂无问题结果，等待审计完成。</div>
          <div v-for="item in findings" :key="item.id || `${item.file}-${item.lineStart}-${item.title}`" class="finding">
            <h3>{{ item.id || '未编号' }}：{{ item.title }}（{{ item.severity }}）</h3>
            <div class="chips">
              <span class="chip">{{ item.category }}</span>
              <span class="chip">{{ item.confidence }}</span>
              <span class="chip">{{ item.file }}:{{ item.lineStart }}-{{ item.lineEnd }}</span>
            </div>
            <p><strong>详细描述：</strong>{{ item.description }}</p>
            <p><strong>影响分析：</strong>{{ item.impact }}</p>
            <p><strong>证据：</strong>{{ item.evidence }}</p>
            <p><strong>修复建议：</strong>{{ item.recommendation }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
