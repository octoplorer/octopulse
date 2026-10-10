export interface ComboboxOption {
  value: string
  label: string
  disabled?: boolean
  group?: string
  description?: string
}
export type ComboboxItem = string | ComboboxOption
