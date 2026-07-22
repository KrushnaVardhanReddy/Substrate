ALTER TABLE contracts
  ADD COLUMN encrypted_content BYTEA,
  ADD COLUMN is_encrypted BOOLEAN DEFAULT false,
  ADD COLUMN kms_key_arn VARCHAR(255);
ALTER TABLE contracts ALTER COLUMN raw_content DROP NOT NULL;
