<script setup lang="ts">
import { ref } from 'vue'
import * as UI from '../ui'
import ChartsShowcase from './ChartsShowcase.vue'
import DisplayShowcase from './DisplayShowcase.vue'
import InputsShowcase from './InputsShowcase.vue'
import NavigationShowcase from './NavigationShowcase.vue'

const dark = ref(false)
const target = ref('https://api.example.com/health')
const description = ref('生产环境 API 的可用性检查。')
const disabled = ref(false)
const readOnly = ref(false)
const invalid = ref(false)
const focusMode = ref<'container' | 'individual'>('container')
const copied = ref(0)
const range = ref('24h')
const saved = ref('尚未提交')
function theme() {
  dark.value = !dark.value
  document.documentElement.dataset.mode = dark.value ? 'dark' : 'light'
}
</script>

<template>
  <UI.Toasty close-label="关闭预览通知">
    <main class="mx-auto max-w-7xl space-y-10 p-5 md:p-10 [&_section[id]]:scroll-mt-6">
      <UI.PageHeader title="Octopulse 组件预览" description="48 组 Kumo 组件与复合区块 · Vue 3 · Ark UI · UnoCSS · VueECharts">
        <template #actions>
          <UI.Button @click="theme">
            {{ dark ? '浅色' : '深色' }}
          </UI.Button><UI.Link href="https://kumo-ui.com" external>
            Kumo 文档
          </UI.Link>
        </template>
      </UI.PageHeader>
      <UI.Breadcrumbs :items="[{ label: '开发工具', href: '/ui.html' }, { label: '完整组件预览' }]" label="当前位置" />
      <nav aria-label="组件分类" class="flex flex-wrap gap-2">
        <UI.Button as-child>
          <a href="#basic-inputs">基础输入与复制</a>
        </UI.Button><UI.Button as-child>
          <a href="#display-layout">展示与布局</a>
        </UI.Button><UI.Button as-child>
          <a href="#advanced-inputs">选择与日期</a>
        </UI.Button><UI.Button as-child>
          <a href="#navigation-overlays">导航与浮层</a>
        </UI.Button><UI.Button as-child>
          <a href="#charts-flow">图表与流程</a>
        </UI.Button>
      </nav>
      <UI.Banner variant="secondary" title="交互验证" description="使用 Tab / Shift+Tab 检查焦点；菜单和选择器可用方向键与 Enter；Escape 关闭浮层。切换深浅色、禁用、只读和错误状态后继续操作。所有操作仅作用于本预览。" />
      <section id="basic-inputs" aria-label="基础输入与复制" class="space-y-5">
        <UI.Text as="h2" variant="heading2">
          基础输入与复制
        </UI.Text>
        <div class="flex flex-wrap gap-4">
          <UI.Switch v-model="disabled" label="禁用基础输入" /><UI.Switch v-model="readOnly" label="只读基础输入" /><UI.Switch v-model="invalid" label="显示输入错误" />
        </div>
        <UI.Grid variant="2up" gap="sm" class="items-start">
          <UI.LayerCard>
            <UI.LayerCardSecondary>Field、Input、InputArea、InputGroup</UI.LayerCardSecondary><UI.LayerCardPrimary>
              <form class="space-y-4" @submit.prevent="saved = `已提交：${target}`">
                <UI.Field label="检查目标" description="有效的 URL；标签、说明和错误会关联控件。" required :disabled="disabled" :read-only="readOnly" :error="invalid ? '请输入正确的检查目标。' : undefined">
                  <UI.Input v-model="target" type="url" required />
                </UI.Field>
                <UI.Field label="监控描述" :disabled="disabled" :read-only="readOnly">
                  <UI.InputArea v-model="description" auto-resize :min-rows="2" :max-rows="5" />
                </UI.Field>
                <UI.Field label="输入组焦点模式">
                  <UI.Select v-model="focusMode">
                    <UI.SelectTrigger><UI.SelectValue /></UI.SelectTrigger><UI.SelectContent>
                      <UI.SelectItem value="container">
                        整个容器
                      </UI.SelectItem><UI.SelectItem value="individual">
                        独立控件
                      </UI.SelectItem>
                    </UI.SelectContent>
                  </UI.Select>
                </UI.Field>
                <UI.InputGroup :focus-mode="focusMode">
                  <UI.InputGroupAddon><span class="i-lucide-search size-4" aria-hidden="true" /></UI.InputGroupAddon><UI.InputGroupInput v-model="target" :disabled="disabled" :readonly="readOnly" aria-label="输入组检查目标" /><UI.InputGroupButton :disabled="disabled" @click="saved = `已搜索：${target}`">
                    搜索
                  </UI.InputGroupButton>
                </UI.InputGroup>
                <UI.Button type="submit" variant="primary" :disabled="disabled">
                  提交预览表单
                </UI.Button><p class="text-subtle" aria-live="polite">
                  {{ saved }}
                </p>
              </form>
            </UI.LayerCardPrimary>
          </UI.LayerCard>
          <UI.LayerCard>
            <UI.LayerCardSecondary>ClipboardText、InlineCopyText、ButtonGroup、Empty</UI.LayerCardSecondary><UI.LayerCardPrimary class="space-y-5">
              <UI.ClipboardText :text="target" copy-label="复制检查目标" copied-label="检查目标已复制" @copy="copied++" /><UI.InlineCopyText text="monitor-preview-01" copy-label="复制监控标识" copied-label="监控标识已复制" @copy="copied++" /><p class="text-subtle" aria-live="polite">
                已复制 {{ copied }} 次
              </p>
              <UI.ButtonGroup label="统计时间范围">
                <UI.Button v-for="item in ['1h', '24h', '7d']" :key="item" :variant="range === item ? 'primary' : 'secondary'" @click="range = item">
                  {{ item }}
                </UI.Button>
              </UI.ButtonGroup><p class="text-subtle" aria-live="polite">
                当前范围：{{ range }}
              </p>
              <UI.Empty title="尚无监控项" description="可用操作按钮触发通知，复制初始化命令。" command-line="octopulse serve" copy-label="复制初始化命令" copied-label="命令已复制" size="sm">
                <template #icon>
                  <span class="i-lucide-activity size-6 text-subtle" aria-hidden="true" />
                </template><template #actions>
                  <UI.Button variant="primary" @click="UI.toastManager.info({ title: '已点击添加监控' })">
                    添加监控示例
                  </UI.Button>
                </template>
              </UI.Empty>
            </UI.LayerCardPrimary>
          </UI.LayerCard>
        </UI.Grid>
      </section>
      <DisplayShowcase />
      <div id="advanced-inputs">
        <UI.Text as="h2" variant="heading2" class="mb-5 block">
          选择与日期
        </UI.Text><InputsShowcase />
      </div>
      <div id="navigation-overlays">
        <UI.Text as="h2" variant="heading2" class="mb-5 block">
          导航与浮层
        </UI.Text><NavigationShowcase />
      </div>
      <div id="charts-flow">
        <UI.Text as="h2" variant="heading2" class="mb-5 block">
          图表与流程
        </UI.Text><ChartsShowcase />
      </div>
    </main>
  </UI.Toasty>
</template>
