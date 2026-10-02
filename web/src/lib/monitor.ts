import { t } from '../composables/i18n'
import type {
  Monitor,
  TLSConfig,
  ConnectionConfig,
  HTTPConfig,
  TCPConfig,
  DNSConfig,
  HeartbeatConfig,
  CertificateConfig,
} from './types'
export const monitorTypes = [
  { value: 'http', label: 'monitorTypes.httpHttps' },
  { value: 'tcp', label: 'monitorTypes.tcpConnection' },
  { value: 'dns', label: 'monitorTypes.dnsResolution' },
  { value: 'heartbeat', label: 'monitorTypes.heartbeat' },
  { value: 'certificate', label: 'monitorTypes.certificateExpiry' },
]
const emptyTLS = (): TLSConfig => ({
  enabled: false,
  insecureSkipVerify: false,
  serverName: '',
  minVersion: '',
  maxVersion: '',
})
const emptyConnection = (): ConnectionConfig => ({
  proxyUrl: '',
  proxyUsername: '',
  dnsServer: '',
  fixedIp: '',
})
export const emptyHTTP = (): HTTPConfig => ({
  url: '',
  method: 'GET',
  query: [],
  headers: [],
  host: '',
  body: {
    format: 'none',
    text: '',
    base64: '',
    charset: 'utf-8',
    contentType: '',
    fields: [],
    files: [],
  },
  auth: { type: 'none', username: '', header: 'Authorization', prefix: '' },
  tls: emptyTLS(),
  connection: emptyConnection(),
  redirects: { enabled: true, maxHops: 5, scope: 'same-origin' },
  acceptEncoding: 'gzip',
  requestGzip: false,
  responseCharset: '',
  maxResponseBytes: 2097152,
  assertions: {
    statusCodes: [],
    statusRanges: [{ min: 200, max: 299 }],
    headers: [],
    textContains: [],
    textNotContains: [],
    regex: [],
    json: [],
    maxLatencyMs: 0,
  },
})
export const emptyTCP = (): TCPConfig => ({
  host: '',
  port: 443,
  tls: emptyTLS(),
  connection: emptyConnection(),
  sendText: '',
  sendBase64: '',
  charset: 'utf-8',
  receiveContains: '',
  receiveRegex: '',
  maxReceiveBytes: 65536,
})
export const emptyDNS = (): DNSConfig => ({
  name: '',
  recordType: 'A',
  server: '',
  protocol: 'udp',
  expectedRCode: 'NOERROR',
  expectedValues: [],
  matchMode: 'contains',
})
export const emptyHeartbeat = (): HeartbeatConfig => ({
  periodSeconds: 60,
  graceSeconds: 30,
  lastReceivedAt: 0,
  lastSuccess: false,
  description: '',
})
export const emptyCertificate = (): CertificateConfig => ({
  host: '',
  port: 443,
  tls: emptyTLS(),
  connection: emptyConnection(),
  warningDays: [30, 14, 7, 1],
  notifyRenewal: true,
  state: 'check_failed',
  expiresAt: 0,
  fingerprint: '',
  daysRemaining: 0,
})
export function newMonitor(): Monitor {
  return {
    id: '',
    name: '',
    description: '',
    type: 'http',
    tags: [],
    group: '',
    enabled: true,
    intervalSeconds: 60,
    timeoutSeconds: 10,
    retries: 2,
    retryDelaySeconds: 1,
    failureThreshold: 1,
    recoveryThreshold: 1,
    notificationChannelIds: [],
    notifyRecovery: true,
    reminderSeconds: 0,
    configVersion: 0,
    state: 'unknown',
    failureCount: 0,
    successCount: 0,
    lastCheckedAt: 0,
    nextCheckAt: 0,
    createdAt: 0,
    updatedAt: 0,
    http: emptyHTTP(),
  }
}
export function targetOf(m: Monitor) {
  return (
    m.http?.url ||
    (m.tcp ? `${m.tcp.host}:${m.tcp.port}` : '') ||
    m.dns?.name ||
    (m.certificate ? `${m.certificate.host}:${m.certificate.port}` : '') ||
    t('monitorTypes.heartbeat')
  )
}
