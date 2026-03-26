export interface StartReviewRequest {
  mrUrl: string
  repoUrl: string
  localRepoPath: string
  sourceBranch: string
  targetBranch: string
  configPath: string
  workspaceDir: string
}

export interface ProgressEvent {
  taskId: string
  stage: string
  percent: number
  message: string
  high: number
  medium: number
  low: number
}

export interface Finding {
  id: string
  title: string
  severity: string
  category: string
  confidence: string
  file: string
  lineStart: number
  lineEnd: number
  description: string
  impact: string
  evidence: string
  recommendation: string
}

export interface Summary {
  highRisk: number
  severe: number
  general: number
  suggestion: number
  high: number
  medium: number
  low: number
  total: number
}

export interface HealthScore {
  security: number
  performance: number
  robustness: number
  maintainability: number
  frameworkPractice: number
}

export interface Report {
  taskId: string
  title: string
  createdAt: string
  findings: Finding[]
  summary: Summary
  health: HealthScore
  notes: string[]
}

export interface ReviewDoneEvent {
  taskId: string
  reportDir: string
  htmlPath: string
  markdownPath: string
  jsonPath: string
  report: Report
}

export interface HistoryItem {
  taskId: string
  title: string
  repoUrl: string
  sourceRef: string
  targetRef: string
  createdAt: string
  reportDir: string
  totalIssues: number
  summary: Summary
}
