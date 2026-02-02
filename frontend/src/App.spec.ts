import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import App from './App.vue'

describe('App', () => {
  it('renders the dashboard title', () => {
    const wrapper = mount(App)
    expect(wrapper.text()).toContain('Dash - Monitoring Dashboard')
  })

  it('renders the description', () => {
    const wrapper = mount(App)
    expect(wrapper.text()).toContain('Grafana-like monitoring dashboard')
  })
})
