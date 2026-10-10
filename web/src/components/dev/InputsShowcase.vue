<script setup lang="ts">
import type { DatePickerPreset, TagInputLabels } from '../ui'
import { computed, ref } from 'vue'
import * as UI from '../ui'

const disabled = ref(false)
const readOnly = ref(false)
const showError = ref(false)
const state = computed(() => ({ disabled: disabled.value, readOnly: readOnly.value }))
const error = computed(() => showError.value ? '这是用于测试的错误提示。' : undefined)
const checked = ref<boolean | 'indeterminate'>('indeterminate')
const channels = ref(['email'])
const interval = ref('60')
const database = ref('postgres')
const databases = ref(['postgres', 'sqlite'])
const combo = ref('postgres')
const combos = ref(['postgres'])
const freeChoice = ref('')
const freeText = ref('custom.example.com')
const hostname = ref('api.example.com')
const tags = ref(['production', 'api'])
const tagText = ref('')
const emails = ref(['ops@example.com'])
const emailText = ref('')
const sensitive = ref('demo-preview-key')
const visible = ref(false)
const copied = ref(0)
const threshold = ref(35)
const range = ref([20, 80])
const dates = ref(['2026-10-09'])
const multipleDates = ref(['2026-10-09', '2026-10-12'])
const rangeDates = ref(['2026-10-09', '2026-10-12'])
const calendarRangeDates = ref(['2026-10-09', '2026-10-12'])
const unavailableDay = ref(false)
const datePresets: DatePickerPreset[] = [
  { label: '发布日', value: ['2026-10-09'] },
  { label: '发布与维护', value: ['2026-10-09', '2026-10-12'] },
]
const rangePresets: DatePickerPreset[] = [
  { label: '本次维护', value: ['2026-10-09', '2026-10-12'] },
  { label: '下次维护', value: ['2026-10-20', '2026-10-23'] },
]
const items = [
  { value: 'postgres', label: 'PostgreSQL', group: '关系数据库', description: '生产数据' },
  { value: 'sqlite', label: 'SQLite', group: '关系数据库', description: '本地开发' },
  { value: 'redis', label: 'Redis', group: '缓存' },
  { value: 'unavailable', label: '暂不可用', group: '缓存', disabled: true },
]
const tagLabels: TagInputLabels = {
  input: '输入标签',
  removeValue: value => `移除标签 ${value}`,
  editValue: value => `编辑标签 ${value}`,
  invalidValue: value => `“${value}”不是有效邮箱。`,
  maxValuesReached: max => `最多输入 ${max} 个值。`,
}
const snapshot = computed(() => JSON.stringify({
  checked: checked.value,
  channels: channels.value,
  interval: interval.value,
  select: database.value,
  multiSelect: databases.value,
  combobox: combo.value,
  multiCombobox: combos.value,
  freeText: freeText.value,
  autocomplete: hostname.value,
  tags: tags.value,
  emails: emails.value,
  secretVisible: visible.value,
  threshold: threshold.value,
  range: range.value,
  dates: dates.value,
  multipleDates: multipleDates.value,
  rangeDates: rangeDates.value,
  calendarRangeDates: calendarRangeDates.value,
}, null, 2))
function reset() {
  disabled.value = false
  readOnly.value = false
  showError.value = false
  checked.value = 'indeterminate'
  channels.value = ['email']
  interval.value = '60'
  database.value = 'postgres'
  databases.value = ['postgres', 'sqlite']
  combo.value = 'postgres'
  combos.value = ['postgres']
  freeChoice.value = ''
  freeText.value = 'custom.example.com'
  hostname.value = 'api.example.com'
  tags.value = ['production', 'api']
  tagText.value = ''
  emails.value = ['ops@example.com']
  emailText.value = ''
  sensitive.value = 'demo-preview-key'
  visible.value = false
  copied.value = 0
  threshold.value = 35
  range.value = [20, 80]
  dates.value = ['2026-10-09']
  multipleDates.value = ['2026-10-09', '2026-10-12']
  rangeDates.value = ['2026-10-09', '2026-10-12']
  calendarRangeDates.value = ['2026-10-09', '2026-10-12']
  unavailableDay.value = false
}
</script>

<template>
  <section class="space-y-6" aria-label="输入组件交互预览">
    <UI.Text as="h2" variant="heading2">
      输入组件交互预览
    </UI.Text>
    <div class="flex flex-wrap items-center gap-5 rounded-lg bg-recessed p-4">
      <UI.Switch v-model="disabled" label="禁用全部示例" />
      <UI.Switch v-model="readOnly" label="只读全部示例" />
      <UI.Switch v-model="showError" label="显示错误状态" />
      <UI.Button @click="reset">
        重置示例
      </UI.Button>
    </div>
    <UI.Grid variant="2up" gap="sm" class="items-start">
      <UI.LayerCard>
        <UI.LayerCardSecondary>Checkbox / Radio</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-5">
          <UI.Checkbox v-model="checked" v-bind="state" :invalid="showError" label="混合选择状态" description="初始为部分选中，单击切换到全选。" />
          <UI.CheckboxGroup v-model="channels" v-bind="state" :error="error" label="通知渠道" :max-selected-values="2" description="最多选择两个渠道。">
            <UI.CheckboxItem value="email" label="邮件" />
            <UI.CheckboxItem value="webhook" label="Webhook" />
            <UI.CheckboxItem value="sms" label="短信（不可用）" disabled />
          </UI.CheckboxGroup>
          <UI.CheckboxGroup v-model="channels" v-bind="state" :error="error" label="卡片式通知渠道" appearance="card">
            <UI.CheckboxItem value="email" label="邮件通知" description="发送给值班人员。" />
            <UI.CheckboxItem value="webhook" label="Webhook 通知" description="发送到外部集成。" />
          </UI.CheckboxGroup>
          <UI.RadioGroup v-model="interval" v-bind="state" :invalid="showError" label="检查间隔" orientation="horizontal">
            <UI.RadioItem value="30" label="30 秒" />
            <UI.RadioItem value="60" label="60 秒" />
            <UI.RadioItem value="300" label="5 分钟" />
            <UI.RadioItem value="disabled" label="不可用" disabled />
          </UI.RadioGroup>
          <UI.RadioGroup v-model="interval" v-bind="state" :invalid="showError" label="卡片式检查间隔" appearance="card" control-position="end">
            <UI.RadioItem value="30" label="每 30 秒" description="快速发现故障。" />
            <UI.RadioItem value="60" label="每 60 秒" description="日常监控。" />
          </UI.RadioGroup>
          <UI.Field label="分段式检查间隔" :error="error">
            <UI.RadioGroup v-model="interval" v-bind="state" :invalid="showError" label="分段式检查间隔" appearance="segmented">
              <UI.RadioItem value="30" label="30 秒" />
              <UI.RadioItem value="60" label="60 秒" />
              <UI.RadioItem value="300" label="5 分钟" />
            </UI.RadioGroup>
          </UI.Field>
        </UI.LayerCardPrimary>
      </UI.LayerCard>
      <UI.LayerCard>
        <UI.LayerCardSecondary>Select / Combobox / Autocomplete</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-5">
          <UI.Select v-model="database" v-bind="state" :error="error" label="单选数据库">
            <UI.SelectTrigger><UI.SelectValue placeholder="选择数据库" /></UI.SelectTrigger>
            <UI.SelectContent>
              <UI.SelectItem v-for="item in items" :key="item.value" :value="item.value" :disabled="item.disabled">
                {{ item.label }}
              </UI.SelectItem>
            </UI.SelectContent>
          </UI.Select>
          <UI.Select v-model="databases" v-bind="state" :error="error" label="多选数据库" multiple>
            <UI.SelectTrigger><UI.SelectValue placeholder="选择一个或多个数据库" /></UI.SelectTrigger>
            <UI.SelectContent>
              <UI.SelectItem v-for="item in items" :key="item.value" :value="item.value" :disabled="item.disabled">
                {{ item.label }}
              </UI.SelectItem>
            </UI.SelectContent>
          </UI.Select>
          <UI.Combobox v-model="combo" v-bind="state" :error="error" :items="items" label="搜索数据库" clearable clear-label="清空数据库" empty-label="没有匹配的数据库" placeholder="输入名称筛选" />
          <UI.Combobox v-model="combos" v-bind="state" :error="error" :items="items" label="搜索并多选数据库" multiple clearable clear-label="清空数据库选择" empty-label="没有匹配的数据库" :remove-label="item => `移除 ${item.label}`" placeholder="添加数据库" />
          <UI.Combobox v-model="freeChoice" v-model:input-value="freeText" v-bind="state" :error="error" :items="['api.example.com', 'status.example.com']" label="Combobox 自由文本" free-text clearable clear-label="清空主机名" empty-label="继续使用输入的主机名" />
          <UI.Autocomplete v-model="hostname" v-bind="state" :error="error" :items="['api.example.com', 'status.example.com']" label="Autocomplete 主机名" empty-label="没有匹配的建议" />
        </UI.LayerCardPrimary>
      </UI.LayerCard>
      <UI.LayerCard>
        <UI.LayerCardSecondary>TagInput / SensitiveInput / Slider</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-5">
          <UI.TagInput v-model="tags" v-model:input-value="tagText" v-bind="state" :error="error" :labels="tagLabels" label="可编辑标签" description="输入后按 Enter 添加；双击标签编辑；粘贴逗号或换行分隔的值。" placeholder="添加标签" :max-values="5" />
          <UI.TagInput v-model="emails" v-model:input-value="emailText" v-bind="state" :error="error" :labels="tagLabels" label="邮箱验证与数量限制" description="最多三个邮箱。粘贴有效值与无效值时，保留被拒绝的输入。" placeholder="ops@example.com" :max-values="3" :validate-value="value => /^.+@.+\..+$/.test(value)" />
          <UI.ClipboardText text="alice@example.com, invalid, bob@example.com" copy-label="复制验证用粘贴内容" copied-label="已复制验证内容" />
          <UI.SensitiveInput v-model="sensitive" v-model:visible="visible" v-bind="state" :error="error" label="演示密钥" description="此值是用于预览的示例文本。" show-label="显示演示密钥" hide-label="隐藏演示密钥" copy-label="复制演示密钥" copied-label="已复制演示密钥" @copy="copied++" />
          <p class="text-size-sm text-subtle" role="status">
            {{ visible ? '密钥当前可见' : '密钥当前隐藏' }} · 已复制 {{ copied }} 次
          </p>
          <div class="pb-7">
            <UI.Slider v-model="threshold" v-bind="state" :invalid="showError" label="延迟阈值" :min="0" :max="100" :step="5" :marks="[{ value: 0, label: '0 ms' }, { value: 50, label: '50 ms' }, { value: 100, label: '100 ms' }]" />
          </div>
          <div class="pb-5">
            <UI.Slider v-model="range" v-bind="state" :invalid="showError" label="延迟范围" :min="0" :max="100" :step="5" :min-steps-between-thumbs="2" />
          </div>
        </UI.LayerCardPrimary>
      </UI.LayerCard>
      <UI.LayerCard>
        <UI.LayerCardSecondary>DatePicker / DateRangePicker</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-5">
          <UI.DatePicker v-model="dates" v-bind="state" :invalid="showError" label="单日维护" locale="zh-CN" min="2026-10-01" max="2026-12-31" clearable clear-label="清空维护日期" />
          <UI.DatePicker v-model="multipleDates" v-bind="state" :invalid="showError" label="多日维护与预设" locale="zh-CN" mode="multiple" :max-selected-dates="3" :presets="datePresets" clearable clear-label="清空所有维护日期" />
          <UI.Switch v-model="unavailableDay" label="将 10 月 14 日设为不可用" />
          <UI.DatePicker v-model="calendarRangeDates" v-bind="state" :invalid="showError" label="日历范围与悬停状态" locale="zh-CN" mode="range" :is-date-unavailable="date => unavailableDay && date.toString() === '2026-10-14'" clearable clear-label="清空日历范围" />
          <UI.DateRangePicker v-model="rangeDates" v-bind="state" :invalid="showError" label="维护窗口" locale="zh-CN" :presets="rangePresets" timezone="Asia/Shanghai" clear-label="重置维护窗口" />
          <UI.Button @click="rangeDates = ['2026-10-20', '2026-10-23']">
            从外部更新维护窗口
          </UI.Button>
          <p class="text-size-sm text-subtle">
            在日历中使用方向键移动焦点、Enter 选择日期，Escape 关闭浮层。
          </p>
        </UI.LayerCardPrimary>
      </UI.LayerCard>
    </UI.Grid>
    <UI.Collapsible>
      <UI.CollapsibleTrigger>查看当前受控值</UI.CollapsibleTrigger>
      <UI.CollapsibleContent class="pt-3">
        <UI.Code :code="snapshot" lang="json" />
      </UI.CollapsibleContent>
    </UI.Collapsible>
  </section>
</template>
