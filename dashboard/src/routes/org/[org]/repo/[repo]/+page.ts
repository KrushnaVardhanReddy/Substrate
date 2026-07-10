import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	// Future phase will fetch schema content from API.
	// For now, return dummy data for UI display.
	return {
		repoName: params.repo,
		orgName: params.org,
		schemaContent: `openapi: 3.0.0
info:
  title: Dummy API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Returns a list of users.
      responses:
        '200':
          description: A JSON array of user names
          content:
            application/json:
              schema:
                type: array
                items:
                  type: string
`
	};
};
