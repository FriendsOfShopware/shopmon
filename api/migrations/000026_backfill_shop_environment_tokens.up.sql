-- Shop creation did not supply a token for its initial environment. Repair
-- empty tokens created since migration 22, preserving all configured tokens.
UPDATE environment
SET environment_token = replace(gen_random_uuid()::text || gen_random_uuid()::text, '-', '')
WHERE environment_token = '';
