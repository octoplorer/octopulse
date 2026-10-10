// @vitest-environment happy-dom
import { mount } from '@vue/test-utils'
import { expect, it, vi } from 'vitest'
import { defineComponent, ref } from 'vue'
import { InputGroup, InputGroupButton, InputGroupInput } from './index'

it('shares disabled state with both controls while preserving the input model', async () => {
  const submit = vi.fn()
  const disabled = ref(false)
  const value = ref('api.example.com')
  const wrapper = mount(defineComponent({
    components: { InputGroup, InputGroupButton, InputGroupInput },
    setup: () => ({ disabled, value, submit }),
    template: '<InputGroup :disabled="disabled" focus-mode="individual"><InputGroupInput v-model="value" aria-label="Target" /><InputGroupButton @click="submit">Search</InputGroupButton></InputGroup>',
  }))
  try {
    await wrapper.get('input').setValue('preview.local')
    expect(value.value).toBe('preview.local')
    wrapper.get('button').element.click()
    expect(submit).toHaveBeenCalledOnce()
    disabled.value = true
    await wrapper.vm.$nextTick()
    expect((wrapper.get('input').element as HTMLInputElement).disabled).toBe(true)
    wrapper.get('button').element.click()
    expect(submit).toHaveBeenCalledOnce()
    expect(value.value).toBe('preview.local')
  }
  finally {
    wrapper.unmount()
  }
})
