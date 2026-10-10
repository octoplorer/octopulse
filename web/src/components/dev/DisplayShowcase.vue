<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import * as UI from '../ui'

const destination = ref('overview')
const selected = ref<string[]>([])
const nameWidth = ref(220)
const amount = ref(73)
const pending = ref(false)
const disabled = ref(false)
const code = ref('const monitor = { name: "API Gateway", enabled: true }')
const noticeType = ref<'success' | 'error' | 'info' | 'warning'>('success')
const result = ref('等待操作')
const services = [{ id: 'api', name: 'API Gateway', status: '正常', latency: 42 }, { id: 'auth', name: 'Authentication', status: '正常', latency: 68 }, { id: 'console', name: 'Internal Console', status: '维护', latency: 31 }]
const allSelected = computed(() => selected.value.length === services.length)
function selectRow(id: string, checked: boolean) {
  selected.value = checked ? [...new Set([...selected.value, id])] : selected.value.filter(value => value !== id)
}
let resizeController: AbortController | undefined
function resize(delta: number) {
  nameWidth.value = Math.max(120, Math.min(360, nameWidth.value + delta))
}
function startResize(event: PointerEvent) {
  resizeController?.abort()
  resizeController = new AbortController()
  const start = event.clientX
  const width = nameWidth.value
  window.addEventListener('pointermove', (next) => {
    nameWidth.value = Math.max(120, Math.min(360, width + next.clientX - start))
  }, { signal: resizeController.signal })
  window.addEventListener('pointerup', () => resizeController?.abort(), { once: true, signal: resizeController.signal })
}
onBeforeUnmount(() => resizeController?.abort())
function notice() {
  UI.toastManager.create({ type: noticeType.value, title: '组件预览通知', description: '可关闭，也可使用操作按钮。', action: {
    label: '记录操作',
    onClick: () => {
      result.value = '通知操作已执行'
    },
  } })
}
</script>

<template>
  <section id="display-layout" class="space-y-5" aria-label="展示与布局组件">
    <UI.PageHeader title="展示与布局" description="Button、Badge、Banner、Grid、LayerCard、Surface、Table、Sidebar、Text、Label、Loader、Meter、Code、Toast。" />
    <UI.Grid variant="2up" gap="sm" class="items-start">
      <UI.LayerCard>
        <UI.LayerCardSecondary>按钮变体与反馈</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-4">
          <div class="flex flex-wrap gap-4">
            <UI.Checkbox v-model="pending" label="加载状态" /><UI.Checkbox v-model="disabled" label="禁用状态" />
          </div>
          <div class="flex flex-wrap gap-3">
            <UI.Button v-for="variant in (['primary', 'secondary', 'ghost', 'destructive', 'secondary-destructive', 'outline'] as const)" :key="variant" :variant="variant" :loading="pending" :disabled="disabled" @click="result = `点击了 ${variant}`">
              {{ variant }}
            </UI.Button>
          </div>
          <div class="flex flex-wrap items-center gap-3">
            <UI.Button v-for="size in (['xs', 'sm', 'base', 'lg'] as const)" :key="size" :size="size" @click="result = `尺寸 ${size}`">
              {{ size }}
            </UI.Button><UI.Button shape="circle" aria-label="圆形按钮" @click="result = '圆形按钮'">
              <span class="i-lucide-plus size-4" />
            </UI.Button><UI.LinkButton href="https://kumo-ui.com" target="_blank">
              Kumo 文档
            </UI.LinkButton>
          </div>
          <div class="flex flex-wrap gap-2">
            <UI.Badge v-for="variant in (['primary', 'secondary', 'error', 'warning', 'success', 'info', 'beta', 'outline', 'red', 'green', 'neutral', 'orange', 'purple', 'teal', 'teal-subtle', 'blue'] as const)" :key="variant" :variant="variant">
              {{ variant }}
            </UI.Badge>
          </div>
          <UI.Banner variant="alert" title="可操作的提示" description="按钮可以记录操作结果。">
            <template #actions>
              <UI.BannerAction @click="result = '已处理提示'">
                处理
              </UI.BannerAction>
            </template>
          </UI.Banner>
          <p class="text-size-sm text-subtle" aria-live="polite">
            {{ result }}
          </p>
        </UI.LayerCardPrimary>
      </UI.LayerCard>
      <UI.LayerCard>
        <UI.LayerCardSecondary>文字、加载与通知</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-4">
          <div class="space-y-2">
            <UI.Text variant="heading2" as="h3">
              服务健康摘要
            </UI.Text><UI.Text variant="secondary" as="p">
              语义色、文字层级与等宽内容。
            </UI.Text><UI.Text variant="mono">
              GET /api/v1/monitors
            </UI.Text><div class="flex gap-4">
              <UI.Text variant="success">
                运行正常
              </UI.Text><UI.Text variant="error">
                连接失败
              </UI.Text>
            </div>
          </div>
          <UI.Label tooltip="说明可以通过鼠标或键盘焦点查看。" tooltip-label="查看字段说明">
            带说明的标签
          </UI.Label>
          <div class="flex items-center gap-5">
            <UI.Loader size="sm" label="小尺寸加载" /><UI.Loader label="默认加载" /><UI.Loader size="lg" label="大尺寸加载" /><UI.SkeletonLine class="w-32" />
          </div>
          <UI.Slider v-model="amount" label="资源使用率" /><UI.Meter :value="amount" label="资源使用率" :variant="amount > 85 ? 'danger' : 'brand'" />
          <UI.Select v-model="noticeType" label="通知类型">
            <UI.SelectTrigger><UI.SelectValue /></UI.SelectTrigger><UI.SelectContent>
              <UI.SelectItem v-for="kind in (['success', 'error', 'info', 'warning'] as const)" :key="kind" :value="kind">
                {{ kind }}
              </UI.SelectItem>
            </UI.SelectContent>
          </UI.Select>
          <UI.Button @click="notice">
            显示通知
          </UI.Button>
        </UI.LayerCardPrimary>
      </UI.LayerCard>
    </UI.Grid>
    <UI.Surface class="p-5">
      <UI.Text variant="heading" as="h3">
        Surface 与响应式 Grid
      </UI.Text><UI.Grid variant="1-2-4up" gap="sm" class="mt-4">
        <UI.GridItem v-for="value in 4" :key="value" variant="bordered">
          分区 {{ value }}
        </UI.GridItem>
      </UI.Grid>
    </UI.Surface>
    <UI.LayerCard>
      <UI.LayerCardSecondary>Table：全选、部分选择、固定列、列宽调整</UI.LayerCardSecondary>
      <UI.TableToolbar>
        <span aria-live="polite">已选择 {{ selected.length }} 项</span><UI.Button size="sm" :disabled="!selected.length" @click="selected = []">
          清除选择
        </UI.Button>
      </UI.TableToolbar>
      <UI.TableContainer>
        <UI.Table layout="fixed" class="min-w-140">
          <UI.TableCaption>服务列表，可用名称列右侧的把手拖动或方向键调整列宽。</UI.TableCaption><UI.TableHeader>
            <UI.TableRow>
              <UI.TableCheckHead :model-value="allSelected" :indeterminate="selected.length > 0 && !allSelected" label="全选服务" @update:model-value="selected = $event ? services.map(service => service.id) : []" /><UI.TableHead :style="{ width: `${nameWidth}px` }" class="relative group" sticky="left">
                服务名称<UI.TableResizeHandle label="调整服务名称列宽" @resize="resize" @resize-start="startResize" />
              </UI.TableHead><UI.TableHead>状态</UI.TableHead><UI.TableHead>响应时间</UI.TableHead>
            </UI.TableRow>
          </UI.TableHeader>
          <UI.TableBody>
            <UI.TableRow v-for="service in services" :key="service.id" :selected="selected.includes(service.id)">
              <UI.TableCheckCell :model-value="selected.includes(service.id)" :label="`选择 ${service.name}`" @update:model-value="selectRow(service.id, $event)" /><UI.TableCell sticky="left" class="truncate">
                {{ service.name }}
              </UI.TableCell><UI.TableCell>
                <UI.Badge :variant="service.status === '正常' ? 'success' : 'info'" dot>
                  {{ service.status }}
                </UI.Badge>
              </UI.TableCell><UI.TableCell>{{ service.latency }} ms</UI.TableCell>
            </UI.TableRow>
          </UI.TableBody>
          <UI.TableFooter>
            <UI.TableRow>
              <UI.TableCell :colspan="4">
                共 3 项服务
              </UI.TableCell>
            </UI.TableRow>
          </UI.TableFooter>
        </UI.Table>
      </UI.TableContainer>
    </UI.LayerCard>
    <UI.LayerCard>
      <UI.LayerCardSecondary>Sidebar：折叠、悬停预览、调整宽度与移动端导航</UI.LayerCardSecondary>
      <UI.SidebarProvider contained resizable peekable class="h-72 overflow-hidden">
        <UI.Sidebar>
          <UI.SidebarHeader><strong>Octopulse</strong></UI.SidebarHeader><UI.SidebarContent>
            <UI.SidebarGroup>
              <UI.SidebarGroupLabel>工作空间</UI.SidebarGroupLabel><UI.SidebarMenu>
                <UI.SidebarMenuButton icon="i-lucide-layout-dashboard" tooltip="概览" :active="destination === 'overview'" @click="destination = 'overview'">
                  概览
                </UI.SidebarMenuButton><UI.SidebarMenuButton icon="i-lucide-activity" tooltip="监控项" :active="destination === 'monitors'" @click="destination = 'monitors'">
                  监控项<UI.SidebarMenuBadge>3</UI.SidebarMenuBadge>
                </UI.SidebarMenuButton>
              </UI.SidebarMenu>
            </UI.SidebarGroup>
          </UI.SidebarContent><UI.SidebarFooter><UI.SidebarTrigger aria-label="切换示例侧栏" /><UI.SidebarClose aria-label="关闭示例导航" class="md:hidden" /></UI.SidebarFooter><UI.SidebarResizeHandle />
        </UI.Sidebar>
        <div class="min-w-0 flex-1 space-y-4 p-5">
          <UI.SidebarTrigger aria-label="打开或收起示例导航" /><UI.Text variant="heading" as="h3">
            {{ destination === 'overview' ? '概览' : '监控项' }}
          </UI.Text><p class="text-subtle">
            这是独立的预览工作空间。可使用折叠按钮、悬停菜单或调整侧栏宽度。
          </p>
        </div>
      </UI.SidebarProvider>
    </UI.LayerCard>
    <UI.LayerCard>
      <UI.LayerCardSecondary>Code：编辑源代码、语法高亮与复制</UI.LayerCardSecondary><UI.LayerCardPrimary class="space-y-4">
        <UI.Field label="代码内容">
          <UI.InputArea v-model="code" auto-resize />
        </UI.Field><UI.Code :code="code" lang="ts" line-numbers /><UI.CodeBlock :code="code" lang="ts" title="monitor.ts" line-numbers copy-label="复制示例代码" copied-label="代码已复制" />
      </UI.LayerCardPrimary>
    </UI.LayerCard>
  </section>
</template>
