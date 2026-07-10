import { readable } from 'svelte/store'; export const page = readable({ params: { org: 'myorg' }, url: { pathname: '/org/myorg' } });
