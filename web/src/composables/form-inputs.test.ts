// @vitest-environment happy-dom
import { useForm } from '@tanstack/vue-form'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { defineComponent, h } from 'vue'
import { Input, InputArea } from '../components/ui/input'
import { newMonitor } from '../lib/monitor'
import { useJSONInput, useListInput } from './form-inputs'
import { i18n } from './i18n'

const wrappers: ReturnType<typeof mount>[] = []
afterEach(() => wrappers.splice(0).forEach(wrapper => wrapper.unmount()))

function renderForm() {
  const component = defineComponent({
    emits: ['save'],
    setup(_, { emit }) {
      function defaults() {
        const values = newMonitor()
        values.http!.assertions.statusCodes = [200]
        return values
      }
      const formApi = useForm({
        defaultValues: defaults(),
        validators: {
          onSubmit: ({ value }) => validate(value),
        },
        onSubmit: ({ value }) => emit('save', value),
      })
      const values = formApi.useSelector(state => state.values)
      const statusCodes = useListInput(
        () => values.value.type === 'http' ? values.value.http?.assertions.statusCodes ?? [] : [],
        value => formApi.setFieldValue('http.assertions.statusCodes', value),
        { parseItem: Number },
      )
      const headers = useJSONInput(
        () => values.value.type === 'http' ? values.value.http?.assertions.headers ?? [] : [],
        value => formApi.setFieldValue('http.assertions.headers', value),
        'Headers',
      )
      function validate(value: ReturnType<typeof newMonitor>): string | undefined {
        return value.type === 'http' ? headers.error.value || undefined : undefined
      }
      return () => h('form', {
        onSubmit(event: Event) {
          event.preventDefault()
          void formApi.handleSubmit()
        },
      }, [
        h('select', {
          'aria-label': 'Monitor type',
          'value': values.value.type,
          'onChange': (event: Event) => formApi.setFieldValue('type', (event.target as HTMLSelectElement).value),
        }, [h('option', { value: 'http' }, 'HTTP'), h('option', { value: 'tcp' }, 'TCP')]),
        values.value.type === 'http'
          ? [
              h(Input, { 'aria-label': 'Status codes', 'modelValue': statusCodes.value, 'onUpdate:modelValue': next => statusCodes.value = String(next ?? '') }),
              h(InputArea, { 'aria-label': 'Headers', 'modelValue': headers.text.value, 'onUpdate:modelValue': next => headers.text.value = String(next ?? '') }),
            ]
          : null,
        headers.error.value ? h('p', { role: 'alert' }, headers.error.value) : null,
        h('button', { type: 'button', onClick: () => formApi.reset(defaults()) }, 'Reset'),
      ])
    },
  })
  const wrapper = mount(component, { global: { plugins: [i18n] } })
  wrappers.push(wrapper)
  return wrapper
}

describe('form input lifecycle', () => {
  it('formats committed list values after leaving and returning to the HTTP configuration', async () => {
    const wrapper = renderForm()
    await wrapper.get('input[aria-label="Status codes"]').setValue('200, 404, ')
    expect(wrapper.get<HTMLInputElement>('input').element.value).toBe('200, 404, ')
    await wrapper.get('select').setValue('tcp')
    await wrapper.get('select').setValue('http')
    expect(wrapper.get<HTMLInputElement>('input').element.value).toBe('200, 404')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.emitted('save')?.[0]?.[0]).toMatchObject({ http: { assertions: { statusCodes: [200, 404] } } })
  })

  it('clears invalid JSON drafts on a type change without losing the last valid assertions', async () => {
    const wrapper = renderForm()
    const assertions = [{ name: 'Accept', operator: 'exists', value: 'json' }]
    await wrapper.get('textarea').setValue(JSON.stringify(assertions))
    await wrapper.get('textarea').setValue('[{')
    expect(wrapper.get('[role="alert"]').text()).toContain('Headers')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.emitted('save')).toBeUndefined()
    await wrapper.get('select').setValue('tcp')
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    await wrapper.get('select').setValue('http')
    expect(wrapper.get<HTMLTextAreaElement>('textarea').element.value).toBe(JSON.stringify(assertions, null, 2))
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.emitted('save')?.[0]?.[0]).toMatchObject({ http: { assertions: { headers: assertions } } })
  })

  it('resets unfinished list and JSON drafts when the form receives replacement values', async () => {
    const wrapper = renderForm()
    await wrapper.get('input').setValue('200, 404, ')
    await wrapper.get('textarea').setValue('[{')
    await wrapper.get('button').trigger('click')
    expect(wrapper.get<HTMLInputElement>('input').element.value).toBe('200')
    expect(wrapper.get<HTMLTextAreaElement>('textarea').element.value).toBe('[]')
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  })
})
