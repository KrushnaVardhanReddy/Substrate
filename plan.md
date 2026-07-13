1. **Remove Binary**: Delete `api/api-server`.

2. **Fix Authorization Middleware**: The `enforce.go` endpoint receives a GitHub OAuth token (`token = localStorage.getItem('github_token')`), but `authMW` expects a system JWT. Use an appropriate middleware or skip it if the token is passed for the GitHub API directly.
   - I'll change the router configuration in `api/internal/server/router.go` to use `http.HandlerFunc(handlers.EnforceGlobalHandler())` without `authMW`.

3. **Check Admin Scope in UI**: Update `dashboard/src/routes/org/[org]/settings/+page.svelte` to check if the user is an admin.
   - Assume there's some `user.role === 'admin'` or similar, or I can fetch the user details to check. I will look at how `layout.svelte` handles data. If no role is exposed, I'll add a simple UI check/message based on standard Substrate dashboard patterns.

4. **Verify E2E Flow**:
   - Restart the vite server.
   - Run the playwright script again to ensure we don't get the "Not Found" error, as the Vite proxy was configured but the dev server wasn't restarted.

5. **Verify and Submit**:
   - Run `api` tests.
   - Run `github-app` tests.
   - Pre-commit.
   - Submit.
