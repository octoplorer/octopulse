CREATE TABLE `incidents` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`service_id` integer NOT NULL,
	`start_time` text NOT NULL,
	`end_time` text,
	`status` text NOT NULL,
	`title` text NOT NULL,
	`description` text NOT NULL,
	FOREIGN KEY (`service_id`) REFERENCES `services`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `service_logs` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`service_id` integer NOT NULL,
	`status` text NOT NULL,
	`latency` integer,
	`timestamp` integer DEFAULT (unixepoch('now')) NOT NULL,
	`status_code` integer,
	`error_message` text,
	FOREIGN KEY (`service_id`) REFERENCES `services`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `services` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`type` text DEFAULT 'public',
	`name` text NOT NULL,
	`tags` text NOT NULL,
	`url` text NOT NULL,
	`method` text DEFAULT 'GET' NOT NULL,
	`interval` integer NOT NULL,
	`updated_at` integer DEFAULT (unixepoch('now')) NOT NULL,
	`created_at` integer DEFAULT (unixepoch('now')) NOT NULL
);
