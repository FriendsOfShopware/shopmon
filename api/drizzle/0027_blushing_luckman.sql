-- Add the name column with a temporary default value
ALTER TABLE `deployment` ADD `name` text NOT NULL DEFAULT 'unnamed_deployment';