-- Environments created after the shop→environment rename often stored an empty
-- token because the frontend still sent "shopToken" while the API expected
-- "environmentToken". Generate tokens for any empty values so the bypass header
-- is visible and usable again.
UPDATE environment
SET environment_token = replace(gen_random_uuid()::text || gen_random_uuid()::text, '-', '')
WHERE environment_token = '';
