export interface Incident {
  id: number
  startTime: Temporal.PlainDateTime
  endTime: Temporal.PlainDateTime | null
  title: string
  description: string | null
}
