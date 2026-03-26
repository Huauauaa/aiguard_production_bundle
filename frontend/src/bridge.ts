import type { HistoryItem, StartReviewRequest } from './types'

declare global {
  interface Window {
    go?: any
    runtime?: {
      EventsOn?: (name: string, callback: (payload: any) => void) => void
    }
  }
}

export const backend = {
  async startReview(payload: StartReviewRequest): Promise<string> {
    if (window.go?.main?.App?.StartReview) {
      return await window.go.main.App.StartReview(payload)
    }
    return crypto?.randomUUID?.() ?? `mock-${Date.now()}`
  },

  async cancelReview(taskId: string): Promise<void> {
    if (window.go?.main?.App?.CancelReview) {
      await window.go.main.App.CancelReview(taskId)
    }
  },

  async listHistory(): Promise<HistoryItem[]> {
    if (window.go?.main?.App?.ListHistory) {
      return await window.go.main.App.ListHistory()
    }
    return []
  },

  on(name: string, callback: (payload: any) => void) {
    if (window.runtime?.EventsOn) {
      window.runtime.EventsOn(name, callback)
    }
  }
}
