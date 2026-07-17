import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const org = params.org;
	const repo = params.repo;

	// According to spec, Substrate stores raw schema strings in DB.
	// For OpenAPI specs, we render with @stoplight/elements using raw YAML.
	// Provide a dummy OpenAPI YAML string for testing this functionality.
	const dummyYaml = `openapi: 3.0.0
info:
  title: ${repo} API
  version: 1.0.0
  description: Automatically generated API documentation for ${repo}.
paths:
  /users:
    get:
      summary: Get all users
      responses:
        '200':
          description: A list of users
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

	return { org, repo, yamlString: dummyYaml };
};
