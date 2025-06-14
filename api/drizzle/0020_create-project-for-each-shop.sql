INSERT INTO project (organization_id, name, description, created_at, updated_at)
SELECT organization_id, name, id, created_at, date('now')
FROM shop
WHERE organization_id IS NOT NULL;
--> statement-breakpoint

UPDATE shop
SET project_id = (
    SELECT id FROM project WHERE description = shop.id
);
--> statement-breakpoint
UPDATE project SET description = NULL;

