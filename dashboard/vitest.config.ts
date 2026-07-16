import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
	plugins: [
		svelte({ compilerOptions: { runes: true } })
	],
	test: {
		environment: 'jsdom',
		setupFiles: ['./vitest-setup.ts'],
		include: ['src/**/*.{test,spec}.{js,ts}'],
		alias: {
			'$env/dynamic/public': '/src/__mocks__/$env/dynamic/public.ts',
			'$app/stores': '/src/__mocks__/$app/stores.ts',
			'$app/navigation': '/src/__mocks__/$app/navigation.ts',
			'$lib': '/src/lib'
		}
	},
	resolve: {
		conditions: ['browser', 'development']
	}
});
