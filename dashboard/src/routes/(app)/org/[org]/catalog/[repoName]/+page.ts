import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const org = params.org;
	const repoName = params.repoName;

	// Mock OpenAPI YAML string based on the spec
	const rawYaml = `
openapi: 3.0.0
info:
  title: ${repoName} API
  version: 1.0.0
  description: Automatically generated API documentation for ${repoName}.
paths:
  /health:
    get:
      summary: Health Check
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    example: "up"
  /api/v1/users:
    get:
      summary: Get Users
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: string
                    name:
                      type: string
`;

	return { org, repoName, rawYaml };
};
