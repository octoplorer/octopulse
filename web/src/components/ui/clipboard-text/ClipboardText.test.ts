// @vitest-environment happy-dom
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, expect, it, vi } from 'vitest'
import { ClipboardText } from './index'

afterEach(() => vi.restoreAllMocks())

// Adapted from Kumo's clipboard-text.test.tsx: displayed and copied values can differ.
it('copies the supplied value and announces success after the clipboard resolves', async () => {
  const writeText = vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue(undefined)
  const wrapper = mount(ClipboardText, { props: { text: 'Visible token', textToCopy: 'actual-token', copyLabel: 'Copy token', copiedLabel: 'Copied' } })
  await wrapper.get('button').trigger('click')
  await flushPromises()
  expect(writeText).toHaveBeenCalledWith('actual-token')
  expect(wrapper.emitted('copy')).toEqual([[]])
  expect(wrapper.get('[aria-live="polite"]').text()).toBe('Copied')
  wrapper.unmount()
})
