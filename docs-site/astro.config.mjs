// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	integrations: [
		starlight({
			title: 'My Docs',
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/withastro/starlight' }],
			sidebar: [
				{
					label: 'Tutorials',
					items: [
						{ label: 'Getting Started', slug: 'tutorials/getting-started' },
					],
				},
				{
					label: 'How-To Guides',
					items: [
						{ label: 'Intentional Breaking Changes', slug: 'guides/intentional-breaking-changes' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'Configuration Reference', slug: 'reference/configuration' },
						{ label: 'CLI Reference', slug: 'reference/cli' },
						{ label: 'Rule Reference', slug: 'reference/rules' },
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
		}),
	],
});
