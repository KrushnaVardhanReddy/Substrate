ALTER TABLE organizations
    DROP COLUMN stripe_api_key_encrypted,
    DROP COLUMN salesforce_url_encrypted,
    DROP COLUMN salesforce_token_encrypted,
    DROP COLUMN salesforce_client_id_encrypted,
    DROP COLUMN salesforce_client_secret_encrypted,
    DROP COLUMN salesforce_username_encrypted,
    DROP COLUMN salesforce_password_encrypted;
