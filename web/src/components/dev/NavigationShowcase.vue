<script setup lang="ts">
import type { CommandPaletteItemData, PaginationLabels } from '../ui/index'
import { computed, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue'
import * as UI from '../ui/index'

const menuAction = ref('尚未执行操作')
const emailEnabled = ref(true)
const interval = ref('30')
const view = ref('health')
const paused = ref(false)
const checks = ref(0)
const toolbarQuery = ref('api')
const popoverOpen = ref(false)
const channelDraft = ref('值班群')
const channel = ref('值班群')
const popoverNotifications = ref(true)
const commandOpen = ref(false)
const commandQuery = ref('')
const commandResult = ref('尚未执行命令')
const notes = ref(0)
const page = ref(1)
const pageSize = ref(10)
const dropdownPages = ref(true)
const cursorPage = ref(1)
const deleteOpen = ref(false)
const failFirst = ref(false)
const deleteAttempts = ref(0)
const deleteResult = ref('尚未开始删除')
const dialogOpen = ref(false)
const dialogSaving = ref(false)
const scheduleName = ref('生产环境探测')
const savedSchedule = ref('生产环境探测')
const scheduleEnabled = ref(true)
const dialogResult = ref('尚未保存配置')
const basicDialogOpen = ref(false)
const basicDialogDraft = ref('核心 API')
const basicDialogName = ref('核心 API')
const basicDialogResult = ref('尚未编辑名称')
const basicDialogFocus = ref('关闭弹窗后可检查焦点恢复')
const basicDialogEditing = ref(false)
const dialogDatabase = ref('SQLite')
const dialogRegion = ref('asia')
const dialogDate = ref(['2026-10-09'])
const segmentedTab = ref('overview')
const lineTab = ref('overview')
const refreshLoading = ref(false)
const refreshes = ref(0)
const tabItems = [{ value: 'overview', label: '概览' }, { value: 'reports', label: '报告（禁用）', disabled: true }, { value: 'events', label: '事件' }]
const anchorContainer = useTemplateRef<HTMLElement>('anchorContainer')
const sections = [
  { id: 'navigation-overview', label: '探测概览', description: '这个示例只在预览中更新本地状态，可以反复操作。' },
  { id: 'navigation-regions', label: '探测区域', description: '滚动这个区域时，目录会同步高亮当前可见章节。' },
  { id: 'navigation-alerts', label: '告警策略', description: '点击目录锚点可以定位到本节，也可以用键盘访问链接。' },
]
const { activeId, selectSection } = UI.useTableOfContentsActiveId({ ids: sections.map(section => section.id), root: () => anchorContainer.value })
const activeSection = computed(() => sections.find(section => section.id === activeId.value)?.label ?? '探测概览')
const paginationLabels: PaginationLabels = { navigation: '监控分页', firstPage: '第一页', previousPage: '上一页', nextPage: '下一页', lastPage: '最后一页', pageNumber: '页码', pageSize: '每页数量', perPage: '每页：' }
const commands: CommandPaletteItemData[] = [
  { value: 'refresh', label: '立即检查监控', description: '增加当前示例的检查次数', group: '监控操作', shortcut: '↵', keywords: ['检查', '刷新'] },
  { value: 'email', label: '切换邮件通知', description: '同步更新菜单中的邮件开关', group: '监控操作', keywords: ['邮件', '通知'] },
  { value: 'note', label: '创建事件备注', description: '记录一条本地示例备注', group: '协作操作', shortcut: '⌘ ↵', keywords: ['备注', '事件'] },
  { value: 'billing', label: '修改计费设置（不可用）', description: '禁用项目不会被键盘选中', group: '协作操作', disabled: true },
]
const actionLabels: Record<string, string> = { check: '已立即检查', export: '已导出示例数据', copy: '已复制示例链接', archive: '已归档示例记录' }
function selectMenu(value: string) {
  if (value === 'check')
    checks.value++
  if (actionLabels[value])
    menuAction.value = actionLabels[value]
}
function selectCommand(item: CommandPaletteItemData, options: { newTab: boolean }) {
  if (item.value === 'refresh')
    checks.value++
  if (item.value === 'email')
    emailEnabled.value = !emailEnabled.value
  if (item.value === 'note')
    notes.value++
  commandResult.value = `已执行：${item.label}${options.newTab ? '（新标签页意图）' : ''}`
}
function commandShortcut(event: KeyboardEvent) {
  if (event.key.toLowerCase() === 'k' && (event.metaKey || event.ctrlKey)) {
    event.preventDefault()
    commandOpen.value = !commandOpen.value
  }
}
function goToSection(id: string) {
  selectSection(id)
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
}
function openDelete(withFailure: boolean) {
  failFirst.value = withFailure
  deleteAttempts.value = 0
  deleteResult.value = withFailure ? '首次请求会失败，保留弹窗以便重试' : '等待确认删除'
  deleteOpen.value = true
}
async function deleteResource() {
  deleteAttempts.value++
  deleteResult.value = `正在进行第 ${deleteAttempts.value} 次删除请求…`
  await new Promise(resolve => setTimeout(resolve, 650))
  if (failFirst.value && deleteAttempts.value === 1) {
    deleteResult.value = '首次请求失败，可以在弹窗中重试'
    throw new Error('模拟网络错误：请再次点击删除重试。')
  }
  deleteResult.value = `删除成功，共请求 ${deleteAttempts.value} 次；弹窗已关闭`
}
async function saveSchedule() {
  if (dialogSaving.value)
    return
  dialogSaving.value = true
  await new Promise(resolve => setTimeout(resolve, 500))
  savedSchedule.value = scheduleName.value
  dialogResult.value = `已保存「${savedSchedule.value}」，计划${scheduleEnabled.value ? '启用' : '暂停'}`
  dialogSaving.value = false
  dialogOpen.value = false
}
function saveChannel() {
  channel.value = channelDraft.value
  popoverOpen.value = false
}
function openBasicDialog(event: MouseEvent) {
  ;(event.currentTarget as HTMLButtonElement).focus()
  basicDialogDraft.value = basicDialogName.value
  basicDialogResult.value = '正在编辑名称，尚未保存'
  basicDialogFocus.value = '正在打开弹窗'
  basicDialogEditing.value = true
  basicDialogOpen.value = true
}
function saveBasicDialog() {
  basicDialogName.value = basicDialogDraft.value.trim()
  basicDialogResult.value = `已保存名称「${basicDialogName.value}」`
  basicDialogEditing.value = false
  basicDialogOpen.value = false
}
async function refreshPreview() {
  if (refreshLoading.value)
    return
  refreshLoading.value = true
  await new Promise(resolve => setTimeout(resolve, 600))
  refreshes.value++
  checks.value++
  refreshLoading.value = false
}
watch(basicDialogOpen, (open) => {
  if (!open && basicDialogEditing.value) {
    basicDialogResult.value = `已关闭，草稿未保存；名称仍为「${basicDialogName.value}」`
    basicDialogEditing.value = false
  }
})
onMounted(() => window.addEventListener('keydown', commandShortcut))
onBeforeUnmount(() => window.removeEventListener('keydown', commandShortcut))
</script>

<template>
  <section aria-labelledby="navigation-showcase-title" class="space-y-4">
    <div>
      <h2 id="navigation-showcase-title" class="text-size-xl font-semibold text-strong">
        导航与浮层交互
      </h2>
      <p class="mt-1 text-size-sm text-subtle">
        这些示例会显示操作结果。可使用 Tab、方向键、Enter 和 Escape 检查键盘交互。
      </p>
    </div>
    <UI.Grid variant="2up" gap="sm" class="items-start">
      <UI.LayerCard>
        <UI.LayerCardSecondary>菜单、单选、多选与子菜单</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-4">
          <UI.DropdownMenu @select="selectMenu">
            <UI.DropdownMenuTrigger as-child>
              <UI.Button>监控操作 <span class="i-lucide-chevron-down size-4" aria-hidden="true" /></UI.Button>
            </UI.DropdownMenuTrigger>
            <UI.DropdownMenuContent>
              <UI.DropdownMenuGroup>
                <UI.DropdownMenuLabel>常用操作</UI.DropdownMenuLabel>
                <UI.DropdownMenuItem value="check">
                  立即检查 <UI.DropdownMenuShortcut>↵</UI.DropdownMenuShortcut>
                </UI.DropdownMenuItem>
                <UI.DropdownMenuItem value="unavailable" disabled>
                  批量编辑（不可用）
                </UI.DropdownMenuItem>
              </UI.DropdownMenuGroup>
              <UI.DropdownMenuSeparator />
              <UI.DropdownMenuCheckboxItem v-model:checked="emailEnabled" value="email">
                邮件通知
              </UI.DropdownMenuCheckboxItem>
              <UI.DropdownMenuGroup>
                <UI.DropdownMenuLabel>检查频率</UI.DropdownMenuLabel>
                <UI.DropdownMenuRadioGroup v-model="interval">
                  <UI.DropdownMenuRadioItem value="30">
                    每 30 秒
                  </UI.DropdownMenuRadioItem>
                  <UI.DropdownMenuRadioItem value="60">
                    每 60 秒
                  </UI.DropdownMenuRadioItem>
                </UI.DropdownMenuRadioGroup>
              </UI.DropdownMenuGroup>
              <UI.DropdownMenuSeparator />
              <UI.DropdownMenuSub placement="right-start" @select="selectMenu">
                <UI.DropdownMenuSubTrigger>更多操作</UI.DropdownMenuSubTrigger>
                <UI.DropdownMenuContent>
                  <UI.DropdownMenuItem value="export">
                    导出数据
                  </UI.DropdownMenuItem>
                  <UI.DropdownMenuItem value="copy">
                    复制链接
                  </UI.DropdownMenuItem>
                  <UI.DropdownMenuItem value="archive" destructive>
                    归档记录
                  </UI.DropdownMenuItem>
                </UI.DropdownMenuContent>
              </UI.DropdownMenuSub>
            </UI.DropdownMenuContent>
          </UI.DropdownMenu>
          <p role="status" class="text-size-sm text-subtle">
            {{ menuAction }} · 邮件{{ emailEnabled ? '开启' : '关闭' }} · 每 {{ interval }} 秒检查 · 已检查 {{ checks }} 次
          </p>
        </UI.LayerCardPrimary>
      </UI.LayerCard>

      <UI.LayerCard>
        <UI.LayerCardSecondary>MenuBar 与工具栏</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-4">
          <UI.MenuBar v-model="view" label="监控视图">
            <UI.MenuBarItem value="health">
              健康状态
            </UI.MenuBarItem>
            <UI.MenuBarItem value="latency">
              响应延迟
            </UI.MenuBarItem>
            <UI.MenuBarItem value="logs" disabled>
              日志（不可用）
            </UI.MenuBarItem>
            <template #content>
              <UI.MenuBarPanel value="health">
                当前视图：所有探测区域均正常。
              </UI.MenuBarPanel>
              <UI.MenuBarPanel value="latency">
                当前视图：平均响应延迟 42 毫秒。
              </UI.MenuBarPanel>
            </template>
          </UI.MenuBar>
          <UI.Toolbar label="监控工具栏" class="flex-wrap">
            <UI.ToolbarButton :pressed="paused" @click="paused = !paused">
              {{ paused ? '恢复' : '暂停' }}
            </UI.ToolbarButton>
            <UI.ToolbarButton disabled>
              编辑
            </UI.ToolbarButton>
            <UI.ToolbarButton @click="checks++">
              检查
            </UI.ToolbarButton>
            <UI.ToolbarInputGroup>
              <span class="i-lucide-search ml-2 size-4 text-subtle" aria-hidden="true" />
              <UI.ToolbarInput v-model="toolbarQuery" label="工具栏搜索" placeholder="搜索监控项" class="w-28" />
            </UI.ToolbarInputGroup>
            <UI.ToolbarLink href="#navigation-overview" @click.prevent="goToSection('navigation-overview')">
              帮助
            </UI.ToolbarLink>
          </UI.Toolbar>
          <p class="text-size-xs text-subtle">
            工具栏用 ← →、Home、End 移动焦点，跳过禁用项；输入框保留光标操作。
          </p>
          <p role="status" class="text-size-sm text-subtle">
            探测{{ paused ? '已暂停' : '运行中' }} · 搜索：{{ toolbarQuery || '无' }} · 已检查 {{ checks }} 次
          </p>
        </UI.LayerCardPrimary>
      </UI.LayerCard>

      <UI.LayerCard>
        <UI.LayerCardSecondary>浮层表单与提示</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-4">
          <div class="flex flex-wrap items-center gap-2">
            <UI.Popover v-model:open="popoverOpen">
              <UI.PopoverTrigger as-child>
                <UI.Button>通知设置</UI.Button>
              </UI.PopoverTrigger>
              <UI.PopoverContent class="w-80 space-y-4">
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <UI.PopoverTitle>通知设置</UI.PopoverTitle>
                    <UI.PopoverDescription>保存后更新下方状态。</UI.PopoverDescription>
                  </div>
                  <UI.PopoverClose as-child label="关闭通知设置">
                    <UI.Button variant="ghost" shape="square" size="sm">
                      <span class="i-lucide-x size-4" aria-hidden="true" />
                    </UI.Button>
                  </UI.PopoverClose>
                </div>
                <UI.Field label="通知频道">
                  <UI.Input v-model="channelDraft" />
                </UI.Field>
                <UI.Switch v-model="popoverNotifications" label="发送故障通知" description="关闭时仅记录故障事件。" />
                <div class="flex justify-end gap-2">
                  <UI.PopoverClose as-child label="取消通知设置">
                    <UI.Button>取消</UI.Button>
                  </UI.PopoverClose>
                  <UI.Button variant="primary" @click="saveChannel">
                    保存
                  </UI.Button>
                </div>
              </UI.PopoverContent>
            </UI.Popover>
            <UI.TooltipProvider :open-delay="150">
              <UI.Tooltip content="可以通过鼠标悬停或键盘聚焦查看说明；Escape 会关闭提示。">
                <UI.Button variant="ghost" shape="square" aria-label="通知帮助">
                  <span class="i-lucide-circle-help size-4" aria-hidden="true" />
                </UI.Button>
              </UI.Tooltip>
            </UI.TooltipProvider>
          </div>
          <p role="status" class="text-size-sm text-subtle">
            当前频道：{{ channel || '未设置' }} · 故障通知{{ popoverNotifications ? '开启' : '关闭' }} · 浮层{{ popoverOpen ? '打开' : '关闭' }}
          </p>
          <UI.Collapsible>
            <UI.CollapsibleTrigger as-child>
              <UI.Button variant="ghost" size="sm">
                <UI.CollapsibleIndicator />键盘操作说明
              </UI.Button>
            </UI.CollapsibleTrigger>
            <UI.CollapsibleContent>
              <p class="pt-2 text-size-sm text-subtle">
                打开浮层后焦点进入表单。Escape 关闭浮层后，焦点返回触发按钮。
              </p>
            </UI.CollapsibleContent>
          </UI.Collapsible>
        </UI.LayerCardPrimary>
      </UI.LayerCard>

      <UI.LayerCard>
        <UI.LayerCardSecondary>命令面板与动作反馈</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-4">
          <UI.Button @click="commandOpen = true">
            打开命令面板 <UI.DropdownMenuShortcut>⌘ / Ctrl K</UI.DropdownMenuShortcut>
          </UI.Button>
          <p role="status" class="text-size-sm text-subtle">
            {{ commandResult }} · 备注 {{ notes }} 条 · 已检查 {{ checks }} 次
          </p>
          <p class="text-size-xs text-subtle">
            输入关键词过滤，↑ ↓ 选择，Enter 执行。Ctrl / ⌘ Enter 会记录新标签页意图。
          </p>
          <UI.CommandPalette v-model:open="commandOpen" v-model:query="commandQuery" :items="commands" title="预览命令面板" label="搜索预览命令" placeholder="搜索检查、邮件或备注…" empty-label="没有匹配的命令" @select="selectCommand">
            <template #item="{ item }">
              <span class="i-lucide-terminal size-4 text-subtle" aria-hidden="true" />
              <div class="min-w-0 flex-1">
                <span class="block font-medium">{{ item.label }}</span><span class="block text-size-xs text-subtle">{{ item.description }}</span>
              </div>
              <kbd v-if="item.shortcut" class="font-sans text-size-xs text-subtle">{{ item.shortcut }}</kbd>
            </template>
            <template #footer>
              <span>↑ ↓ 选择 · Enter 执行 · Escape 关闭</span><span>当前搜索：{{ commandQuery || '全部' }}</span>
            </template>
          </UI.CommandPalette>
        </UI.LayerCardPrimary>
      </UI.LayerCard>

      <UI.LayerCard>
        <UI.LayerCardSecondary>分页、页大小与未知总数</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-4">
          <UI.Switch v-model="dropdownPages" label="使用下拉框选择页码" />
          <UI.Pagination v-model:page="page" v-model:page-size="pageSize" :count="97" :labels="paginationLabels">
            <UI.PaginationInfo>
              <template #default="{ start, end, count }">
                显示 {{ start }}–{{ end }} 条，共 {{ count }} 条
              </template>
            </UI.PaginationInfo>
            <UI.PaginationPageSize :options="[5, 10, 25]" />
            <UI.PaginationSeparator />
            <UI.PaginationControls :page-selector="dropdownPages ? 'dropdown' : 'input'" />
          </UI.Pagination>
          <p role="status" class="text-size-sm text-subtle">
            当前第 {{ page }} 页，每页 {{ pageSize }} 条；更改页大小会返回第 1 页。
          </p>
          <div class="space-y-2 border-t border-line pt-4">
            <p class="text-size-xs text-subtle">
              未知总数示例：第 3 页后没有下一页。
            </p>
            <UI.Pagination v-model:page="cursorPage" :page-size="10" :has-next-page="cursorPage < 3" :labels="paginationLabels" :show-info="false" controls="simple" />
            <p role="status" class="text-size-sm text-subtle">
              已加载第 {{ cursorPage }} 页
            </p>
          </div>
        </UI.LayerCardPrimary>
      </UI.LayerCard>

      <UI.LayerCard>
        <UI.LayerCardSecondary>基础 Dialog：受控状态与焦点恢复</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-4">
          <UI.Button @click="openBasicDialog" @focus="basicDialogFocus = '焦点位于打开弹窗按钮'">
            编辑监控名称
          </UI.Button>
          <p class="text-size-sm text-subtle">
            保存会更新名称；取消、Escape 或关闭按钮会舍弃草稿。关闭后观察打开按钮的焦点。
          </p>
          <p role="status" class="text-size-sm text-subtle">
            {{ basicDialogResult }} · 弹窗{{ basicDialogOpen ? '打开' : '关闭' }} · {{ basicDialogFocus }}
          </p>
          <UI.Dialog v-model:open="basicDialogOpen" title="编辑监控名称" description="验证基础 Dialog 与内部的选择器、日历和浮层，可用鼠标与键盘操作。" size="lg" close-label="关闭名称编辑" @focusin="basicDialogFocus = '焦点位于弹窗内'">
            <div class="space-y-4">
              <UI.Field label="监控名称" description="请输入新的名称，保存后会更新预览中的状态。">
                <UI.Input v-model="basicDialogDraft" />
              </UI.Field>
              <UI.Select v-model="dialogRegion" label="弹窗探测区域">
                <UI.SelectTrigger><UI.SelectValue /></UI.SelectTrigger>
                <UI.SelectContent>
                  <UI.SelectItem value="asia">
                    亚洲
                  </UI.SelectItem><UI.SelectItem value="europe">
                    欧洲
                  </UI.SelectItem>
                </UI.SelectContent>
              </UI.Select>
              <UI.Combobox v-model="dialogDatabase" label="弹窗数据库" :items="['SQLite', 'PostgreSQL', 'Redis']" />
              <UI.DatePicker v-model="dialogDate" label="弹窗开始日期" :inline="false" locale="zh-CN" />
              <UI.Popover>
                <UI.PopoverTrigger as-child>
                  <UI.Button>弹窗内高级设置</UI.Button>
                </UI.PopoverTrigger>
                <UI.PopoverContent>
                  <UI.Field label="弹窗内通知渠道">
                    <UI.Input v-model="channelDraft" />
                  </UI.Field><UI.PopoverClose as-child>
                    <UI.Button class="mt-3">
                      完成设置
                    </UI.Button>
                  </UI.PopoverClose>
                </UI.PopoverContent>
              </UI.Popover>
              <p role="status" class="text-size-sm text-subtle">
                {{ dialogDatabase }} · {{ dialogRegion }} · {{ dialogDate.join(' 至 ') }}
              </p>
            </div>
            <template #footer>
              <UI.Button @click="basicDialogOpen = false">
                取消
              </UI.Button>
              <UI.Button variant="primary" :disabled="!basicDialogDraft.trim()" @click="saveBasicDialog">
                保存名称
              </UI.Button>
            </template>
          </UI.Dialog>
        </UI.LayerCardPrimary>
      </UI.LayerCard>

      <UI.LayerCard>
        <UI.LayerCardSecondary>Tabs：分段式与下划线</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-4">
          <div class="space-y-2">
            <p class="text-size-xs font-medium text-subtle">
              分段式（Kumo 默认）
            </p>
            <UI.Tabs v-model="segmentedTab" :items="tabItems" variant="segmented">
              <template #overview>
                <p class="pt-3 text-size-sm text-default">
                  概览：当前所有探测区域均正常。
                </p>
              </template>
              <template #events>
                <p class="pt-3 text-size-sm text-default">
                  事件：今日收到 3 条状态更新。
                </p>
              </template>
            </UI.Tabs>
          </div>
          <div class="space-y-2">
            <p class="text-size-xs font-medium text-subtle">
              下划线（line）
            </p>
            <UI.Tabs v-model="lineTab" :items="tabItems" variant="line">
              <template #overview>
                <p class="pt-3 text-size-sm text-default">
                  概览：平均延迟为 42 毫秒。
                </p>
              </template>
              <template #events>
                <p class="pt-3 text-size-sm text-default">
                  事件：最近一次检查已经完成。
                </p>
              </template>
            </UI.Tabs>
          </div>
          <p class="text-size-xs text-subtle">
            ← → 切换并跳过禁用项，Home、End 定位首尾；也可点击下方按钮更新受控值。
          </p>
          <UI.Button size="sm" @click="segmentedTab = 'events'; lineTab = 'events'">
            两个示例都切换到事件
          </UI.Button>
          <p role="status" class="text-size-sm text-subtle">
            分段式：{{ segmentedTab === 'overview' ? '概览' : '事件' }} · 下划线：{{ lineTab === 'overview' ? '概览' : '事件' }}
          </p>
          <div class="flex items-center gap-3 border-t border-line pt-4">
            <UI.RefreshButton :loading="refreshLoading" :label="refreshLoading ? '正在刷新预览数据' : '刷新预览数据'" @click="refreshPreview" />
            <span role="status" class="text-size-sm text-subtle">{{ refreshLoading ? '模拟请求中，按钮暂时禁用…' : `已刷新 ${refreshes} 次` }}</span>
          </div>
        </UI.LayerCardPrimary>
      </UI.LayerCard>

      <UI.LayerCard>
        <UI.LayerCardSecondary>确认删除与弹窗插槽</UI.LayerCardSecondary>
        <UI.LayerCardPrimary class="space-y-4">
          <div class="flex flex-wrap gap-2">
            <UI.Button variant="destructive" @click="openDelete(false)">
              演示删除成功
            </UI.Button>
            <UI.Button @click="openDelete(true)">
              演示失败后重试
            </UI.Button>
          </div>
          <p role="status" class="text-size-sm text-subtle">
            {{ deleteResult }}
          </p>
          <UI.DeleteResource v-model:open="deleteOpen" resource-type="监控项" resource-name="demo-api" title="删除示例监控项" description="输入名称后才能确认。请求过程中禁止关闭或重复提交。" confirmation-label="输入 demo-api 以确认" delete-button-text="删除监控项" cancel-label="取消" :on-delete="deleteResource">
            <p class="text-size-sm text-subtle">
              这是本地交互演示。首次失败模式会保留输入内容，第二次提交会成功。
            </p>
          </UI.DeleteResource>
          <UI.LayerDialog v-model:open="dialogOpen" :dismiss-disabled="dialogSaving">
            <UI.LayerDialogTrigger as-child>
              <UI.Button>编辑探测计划</UI.Button>
            </UI.LayerDialogTrigger>
            <UI.LayerDialogContent close-label="关闭探测计划">
              <template #header>
                <UI.LayerDialogTitle>探测计划</UI.LayerDialogTitle>
                <UI.LayerDialogDescription>自定义标题区与底部操作，异步保存期间锁定关闭。</UI.LayerDialogDescription>
              </template>
              <UI.LayerDialogBody class="space-y-4">
                <UI.Field label="计划名称">
                  <UI.Input v-model="scheduleName" :disabled="dialogSaving" />
                </UI.Field>
                <UI.Switch v-model="scheduleEnabled" label="启用计划" :disabled="dialogSaving" />
              </UI.LayerDialogBody>
              <template #footer>
                <UI.LayerDialogActions>
                  <UI.LayerDialogClose>取消</UI.LayerDialogClose>
                  <UI.LayerDialogAction :loading="dialogSaving" :disabled="!scheduleName.trim()" @click="saveSchedule">
                    保存计划
                  </UI.LayerDialogAction>
                </UI.LayerDialogActions>
              </template>
            </UI.LayerDialogContent>
          </UI.LayerDialog>
          <p role="status" class="text-size-sm text-subtle">
            {{ dialogResult }}
          </p>
        </UI.LayerCardPrimary>
      </UI.LayerCard>
    </UI.Grid>

    <UI.LayerCard>
      <UI.LayerCardSecondary>目录锚点与滚动同步</UI.LayerCardSecondary>
      <UI.LayerCardPrimary class="space-y-3">
        <div class="grid gap-4 sm:grid-cols-[12rem_1fr]">
          <UI.TableOfContents label="预览章节目录">
            <UI.TableOfContentsTitle>本节目录</UI.TableOfContentsTitle>
            <UI.TableOfContentsList>
              <UI.TableOfContentsGroup label="探测概览" href="#navigation-overview" :active="activeId === 'navigation-overview'" @click.prevent="goToSection('navigation-overview')">
                <UI.TableOfContentsItem v-for="section in sections.slice(1)" :key="section.id" :href="`#${section.id}`" :active="activeId === section.id" @click.prevent.stop="goToSection(section.id)">
                  {{ section.label }}
                </UI.TableOfContentsItem>
              </UI.TableOfContentsGroup>
            </UI.TableOfContentsList>
          </UI.TableOfContents>
          <div ref="anchorContainer" tabindex="0" aria-label="可滚动的探测说明" class="h-64 overflow-y-auto rounded-lg border border-line bg-tint">
            <section v-for="section in sections" :id="section.id" :key="section.id" class="min-h-48 scroll-mt-2 border-b border-line p-5 last:border-b-0">
              <h3 class="text-size-base font-semibold text-default">
                {{ section.label }}
              </h3>
              <p class="mt-2 text-size-sm text-subtle">
                {{ section.description }}
              </p>
            </section>
          </div>
        </div>
        <p role="status" class="text-size-sm text-subtle">
          当前章节：{{ activeSection }}
        </p>
      </UI.LayerCardPrimary>
    </UI.LayerCard>
  </section>
</template>
