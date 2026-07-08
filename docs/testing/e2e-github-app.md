# Substrate GitHub App: Local End-to-End Testing

This guide walks you through setting up a true end-to-end local testing environment for the Substrate GitHub App. Instead of relying purely on unit tests and mocked HTTP requests, this setup uses **real GitHub Webhooks** routed to your local development environment via `ngrok`.

This is the recommended way to manually verify UI changes (like PR comment formatting) and test integration bugs before deploying to production.

---

## Prerequisites
1. **Go** installed (to run the diff engine).
2. **Node.js (v20+)** installed (to run the Cloudflare Worker).
3. **ngrok** (or `smee.io`) installed to expose your localhost to the internet.
4. Admin access to a throwaway/test GitHub repository.

---

## Step 1: Start the Go Container Service locally

The GitHub App relies on the Go binary to perform the actual diff computation. You need to run it in its HTTP server mode.

Open a terminal and run:
```bash
cd engine
go run ./cmd/substrate serve --port 8080
```
*Leave this running in the background. It will listen for `POST /diff` requests.*

---

## Step 2: Start the Cloudflare Worker locally

Next, you need to start the Cloudflare Worker webhook receiver. We need to tell it to point to your local Go service instead of the production one.

Open a **second terminal** and run:
```bash
cd github-app

# Create local development variables
echo "CONTAINER_SERVICE_URL=http://127.0.0.1:8080" > .dev.vars

# Install dependencies if you haven't already
npm install

# Start the local worker
npm run dev
```
*This starts the worker locally, usually on `http://localhost:8787`.*

---

## Step 3: Expose your Worker to the internet

GitHub needs a public URL to send webhook events to. We will use `ngrok` to create a secure tunnel to your local worker.

Open a **third terminal** and run:
```bash
ngrok http 8787
```
Ngrok will generate a public Forwarding URL (e.g., `https://a1b2c3d4.ngrok.app`). Copy this URL; you will need it in the next step.

---

## Step 4: Create a Temporary Test GitHub App

You need a private GitHub App installed on your test repository that points to your `ngrok` URL.

1. Go to your GitHub profile: **Settings > Developer settings > GitHub Apps > New GitHub App**.
2. **Name**: "Substrate Local Test" (or similar).
3. **Homepage URL**: Any URL (e.g., your repository URL).
4. **Webhook URL**: Paste your `ngrok` URL from Step 3.
5. **Webhook Secret**: Leave this blank for local testing (or set one and add `GITHUB_WEBHOOK_SECRET=your_secret` to your `.dev.vars` file).
6. **Permissions**:
   - **Pull Requests**: `Read & write` (to post comments).
   - **Commit statuses**: `Read & write` (to block merges).
   - **Contents**: `Read-only` (to read the `substrate.yaml` config).
7. **Subscribe to Events**: Check `Pull request`.
8. Click **Create GitHub App**.

### Generate App Credentials
After creating the app, generate a private key:
1. Note your **App ID** at the top of the General settings page.
2. Scroll down and click **Generate a private key**. This will download a `.pem` file to your computer.

### Update `.dev.vars`
Open your `.dev.vars` file in `github-app/` and add these credentials:
```env
CONTAINER_SERVICE_URL=http://127.0.0.1:8080
APP_ID=your_app_id_here
PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
```
*(Note: Replace newlines in your private key with `\n` so it sits on a single line, or let Wrangler handle multiline secrets if using `wrangler secret put`).*

Restart your `npm run dev` command so it picks up the new credentials.

---

## Step 5: Run the End-to-End Test

1. **Install the App**: On your GitHub App's settings page, go to **Install App** and install it on a test repository.
2. **Setup Base**: In your test repository, commit a base `openapi.yaml` file to the `main` branch.
3. **Trigger Webhook**: Create a new branch, introduce a breaking change to the `openapi.yaml` (like deleting a required property), and open a Pull Request.

### What to watch for:
- Your **ngrok terminal** will show an incoming `POST` request.
- Your **Cloudflare Worker terminal** will log the webhook payload and the attempt to reach the container service.
- Your **Go binary terminal** will log the diff execution.
- Finally, check your **GitHub Pull Request**! You should see Substrate post a comment detailing the breaking change and emitting a failing Commit Status.
