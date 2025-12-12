CREATE TABLE `logs` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`monitor_id` integer NOT NULL,
	`status_code` integer,
	`latency` integer,
	`is_fail` integer DEFAULT 0 NOT NULL,
	`reason` text,
	`created_at` text NOT NULL,
	FOREIGN KEY (`monitor_id`) REFERENCES `monitors`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `monitors` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`name` text NOT NULL,
	`url` text NOT NULL,
	`method` text DEFAULT 'GET' NOT NULL,
	`interval` integer NOT NULL,
	`status` text NOT NULL,
	`retry_count` integer DEFAULT 0 NOT NULL,
	`last_check` text,
	`keyword` text,
	`user_agent` text,
	`domain_expiry` text,
	`cert_expiry` text,
	`check_info_status` text,
	`created_at` text NOT NULL
);
