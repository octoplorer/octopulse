export interface Incident {
  id: number
  startTime: Date
  endTime: string | null
  status: 'ongoing' | 'resolved'
  title: string
  description: string | null
}
