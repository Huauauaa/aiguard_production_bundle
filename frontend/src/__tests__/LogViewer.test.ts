import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import LogViewer from '../components/LogViewer.vue'

describe('LogViewer', () => {
  it('renders nothing when show is false', () => {
    const wrapper = mount(LogViewer, {
      props: { show: false, logPath: '', content: '' }
    })
    expect(wrapper.find('.log-viewer').exists()).toBe(false)
  })

  it('renders log content when show is true', () => {
    const wrapper = mount(LogViewer, {
      props: {
        show: true,
        logPath: '/path/to/log',
        content: 'test log content'
      }
    })
    expect(wrapper.find('.log-path').text()).toBe('/path/to/log')
    expect(wrapper.find('.log-content').text()).toBe('test log content')
  })
})
