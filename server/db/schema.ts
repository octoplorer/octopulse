import { relations, sql } from 'drizzle-orm'
import { integer, sqliteTable, text } from 'drizzle-orm/sqlite-core'

// #region [TABLES]
export const services = sqliteTable('services', {
  id: integer('id').primaryKey({ autoIncrement: true }),

  /**
   * Service type
   *
   * - `public`: service will show service url to public
   * - `private`: service will not show service url to public
   *
   * @default 'public'
   */
  type: text('type', { enum: ['public', 'private'] }).default('public'),
  /** Service name */
  name: text('name').notNull(),
  /** Service tags */
  tags: text('tags', { mode: 'json' }).$type<string[]>().notNull(),

  /** Service health check url */
  url: text('url').notNull(),
  /** Service health check method */
  method: text('method', { enum: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'] }).notNull().default('GET'),

  updatedAt: integer('updated_at', { mode: 'timestamp' })
    .default(sql`(unixepoch('now'))`)
    .$onUpdate(() => new Date())
    .notNull(),
  createdAt: integer('created_at', { mode: 'timestamp' })
    .default(sql`(unixepoch('now'))`)
    .notNull(),
  deletedAt: integer('deleted_at', { mode: 'timestamp' }),
})

export const serviceLogs = sqliteTable('service_logs', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  serviceId: integer('service_id')
    .notNull()
    .references(() => services.id),

  status: text('status', { enum: ['up', 'timeout', 'error'] })
    .notNull(),
  latency: integer('latency'),

  timestamp: integer('timestamp', { mode: 'timestamp' })
    .default(sql`(unixepoch('now'))`)
    .notNull(),

  statusCode: integer('status_code'),
  errorMessage: text('error_message'),
  deletedAt: integer('deleted_at', { mode: 'timestamp' }),
})

export const incidents = sqliteTable('incidents', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  serviceId: integer('service_id')
    .references(() => services.id)
    .notNull(),
  startTime: text('start_time').notNull(),
  endTime: text('end_time'),
  status: text('status', { enum: ['ongoing', 'resolved'] }).notNull(),
  title: text('title').notNull(),
  description: text('description').notNull(),
})

// #region [RELATIONS]
export const monitorsRelations = relations(services, ({ many }) => ({
  incidents: many(incidents),
  logs: many(serviceLogs),
}))

export const serviceLogsRelations = relations(serviceLogs, ({ one }) => ({
  service: one(services, {
    fields: [serviceLogs.serviceId],
    references: [services.id],
  }),
}))

export const incidentsRelations = relations(incidents, ({ one }) => ({
  service: one(services, {
    fields: [incidents.serviceId],
    references: [services.id],
  }),
}))
