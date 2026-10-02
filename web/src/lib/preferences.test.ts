import test from 'node:test'
import assert from 'node:assert/strict'
import { datetimeInput, datetimeMilliseconds, duration, formatPercent } from './preferences.ts'

test('maintenance wall-clock input uses the selected zone, not the process zone', () => {
  const timestamp = datetimeMilliseconds('2026-10-02T09:30', 'Asia/Shanghai')
  assert.equal(timestamp, Date.UTC(2026, 9, 2, 1, 30))
  assert.equal(datetimeInput(timestamp, 'Asia/Shanghai'), '2026-10-02T09:30')
  assert.equal(datetimeInput(timestamp, 'UTC'), '2026-10-02T01:30')
})

test('nonexistent spring DST times are rejected', () => {
  assert.throws(() => datetimeMilliseconds('2026-03-08T02:30', 'America/New_York'))
})

test('normal DST time round-trips in the selected zone', () => {
  const timestamp = datetimeMilliseconds('2026-07-01T15:45', 'America/New_York')
  assert.equal(timestamp, Date.UTC(2026, 6, 1, 19, 45))
  assert.equal(datetimeInput(timestamp, 'America/New_York'), '2026-07-01T15:45')
})

test('missing statistics never appear as perfect uptime', () => {
  assert.equal(formatPercent(null), '—')
  assert.equal(formatPercent(0), '0.00%')
  assert.equal(formatPercent(1), '100.00%')
  assert.equal(duration(86400000), '1.0 d')
})
