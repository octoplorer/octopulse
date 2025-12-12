import { sqliteTable, integer, text } from 'drizzle-orm/sqlite-core';
import { relations } from 'drizzle-orm';

// monitors table
export const monitors = sqliteTable('monitors', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  name: text('name').notNull(),
  url: text('url').notNull(),
  method: text('method').notNull().default('GET'),
  interval: integer('interval').notNull(),
  status: text('status').notNull(), // UP/DOWN/RETRYING
  retryCount: integer('retry_count').notNull().default(0),
  lastCheck: text('last_check'), // DATETIME stored as ISO string
  keyword: text('keyword'), // Optional
  userAgent: text('user_agent'), // Optional
  domainExpiry: text('domain_expiry'), // DATETIME stored as ISO string
  certExpiry: text('cert_expiry'), // DATETIME stored as ISO string
  checkInfoStatus: text('check_info_status'),
  createdAt: text('created_at').notNull(), // DATETIME stored as ISO string
});

// logs table
export const logs = sqliteTable('logs', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  monitorId: integer('monitor_id')
    .notNull()
    .references(() => monitors.id, { onDelete: 'cascade' }),
  statusCode: integer('status_code'),
  latency: integer('latency'), // Response time in ms
  isFail: integer('is_fail').notNull().default(0), // BOOLEAN as INTEGER (0/1)
  reason: text('reason'),
  createdAt: text('created_at').notNull(), // DATETIME stored as ISO string
});

// Relations
export const monitorsRelations = relations(monitors, ({ many }) => ({
  logs: many(logs),
}));

export const logsRelations = relations(logs, ({ one }) => ({
  monitor: one(monitors, {
    fields: [logs.monitorId],
    references: [monitors.id],
  }),
}));
