import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import HistoryList from '../components/HistoryList.vue'

describe('HistoryList', () => {
  it('shows empty message when no history', () => {
    const wrapper = mount(HistoryList, {
      props: { history: [] }
    })
    expect(wrapper.find('.empty').text()).toBe('暂无历史记录')
  })

  it('renders history items', () => {
    const wrapper = mount(HistoryList, {
      props: {
        history: [
          {
            taskId: 'abc123',
            timestamp: '2026-03-26 10:00:00',
            sourceBranch: 'feature',
            targetBranch: 'main',
            findingCount: 5
          }
        ]
      }
    })
    expect(wrapper.find('.task-id').text()).toBe('abc123')
    expect(wrapper.find('.finding-count').text()).toBe('5 个问题')
  })
})
