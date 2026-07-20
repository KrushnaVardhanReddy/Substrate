ALTER TABLE contracts
  DROP COLUMN encrypted_content,
  DROP COLUMN is_encrypted,
  DROP COLUMN kms_key_arn;
ALTER TABLE contracts ALTER COLUMN raw_content SET NOT NULL;
