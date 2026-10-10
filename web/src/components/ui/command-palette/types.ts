export interface CommandPaletteItemData {
  value: string
  label: string
  description?: string
  group?: string
  shortcut?: string
  disabled?: boolean
  keywords?: string[]
}
export type HighlightRange = [number, number]
