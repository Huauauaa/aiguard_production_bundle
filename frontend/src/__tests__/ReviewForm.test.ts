import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ReviewForm from '../components/ReviewForm.vue'

describe('ReviewForm', () => {
  const defaultProps = {
    form: {
      mrUrl: '',
      repoUrl: '',
      localRepoPath: '',
      sourceBranch: '',
      targetBranch: '',
      configPath: 'config.yaml',
      workspaceDir: './workspace'
    },
    availableBranches: [],
    running: false,
    pullingCode: false,
    clearingCache: false,
    showLogs: false,
    disabled: false
  }

  it('renders form fields', () => {
    const wrapper = mount(ReviewForm, { props: defaultProps })
    expect(wrapper.find('input[placeholder*="MR"]').exists()).toBe(true)
  })

  it('emits pull-code event', async () => {
    const wrapper = mount(ReviewForm, { props: defaultProps })
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('pull-code')).toBeTruthy()
  })

  it('disables buttons when disabled prop is true', () => {
    const wrapper = mount(ReviewForm, {
      props: { ...defaultProps, disabled: true }
    })
    const buttons = wrapper.findAll('button')
    buttons.forEach(btn => {
      expect(btn.attributes('disabled')).toBeDefined()
    })
  })
})
