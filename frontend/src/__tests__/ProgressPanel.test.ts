import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ProgressPanel from '../components/ProgressPanel.vue'

describe('ProgressPanel', () => {
  it('renders nothing when progress is null', () => {
    const wrapper = mount(ProgressPanel, {
      props: { progress: null }
    })
    expect(wrapper.find('.progress-panel').exists()).toBe(false)
  })

  it('renders progress when provided', () => {
    const wrapper = mount(ProgressPanel, {
      props: {
        progress: {
          stage: '扫描中',
          percent: 50,
          message: '正在扫描文件'
        }
      }
    })
    expect(wrapper.find('.stage').text()).toBe('扫描中')
    expect(wrapper.find('.percent').text()).toBe('50%')
    expect(wrapper.find('.progress-message').text()).toBe('正在扫描文件')
  })
})
