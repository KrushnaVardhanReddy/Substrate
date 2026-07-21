ALTER TABLE organizations
    ADD COLUMN stripe_api_key_encrypted TEXT,
    ADD COLUMN salesforce_url_encrypted TEXT,
    ADD COLUMN salesforce_token_encrypted TEXT,
    ADD COLUMN salesforce_client_id_encrypted TEXT,
    ADD COLUMN salesforce_client_secret_encrypted TEXT,
    ADD COLUMN salesforce_username_encrypted TEXT,
    ADD COLUMN salesforce_password_encrypted TEXT;
