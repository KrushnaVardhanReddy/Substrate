import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
	integrations: [
		starlight({
			title: 'Substrate',
			social: {
				github: 'https://github.com/KrushnaVardhanReddy/Substrate',
			},
			sidebar: [
				{
					label: 'Tutorials',
					items: [{ label: 'Getting Started', slug: 'tutorials/getting-started' }],
				},
				{
					label: 'How-To Guides',
					items: [
						{ label: 'Intentional Breaking Changes', slug: 'guides/intentional-breaking-changes' },
						{ label: 'MCP & AI IDE Integration', slug: 'guides/mcp-ide-integration' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'Configuration', slug: 'reference/configuration' },
						{ label: 'CLI', slug: 'reference/cli' },
						{ label: 'Rules', slug: 'reference/rules' },
					],
				},
				{
					label: 'Explanation',
					items: [
						{ label: 'Contract Registry', slug: 'explanation/contract-registry' },
						{ label: 'AI & Shift-Left', slug: 'explanation/ai-shift-left' },
					],
				},
			],
			customCss: ['./src/tailwind.css'],
		}),
	],
});