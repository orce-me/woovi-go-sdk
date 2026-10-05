// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import { ion } from 'starlight-ion-theme';

// https://astro.build/config
export default defineConfig({
	site: 'https://orce-me.github.io',
	base: '/woovi-go-sdk',
	integrations: [
		starlight({
			title: 'woovi-go-sdk',
			description: 'Go SDK for the Woovi (OpenPix) REST API',
			social: [
				{
					icon: 'github',
					label: 'GitHub',
					href: 'https://github.com/orce-me/woovi-go-sdk',
				},
			],
			editLink: {
				baseUrl: 'https://github.com/orce-me/woovi-go-sdk/edit/main/docs/',
			},
			lastUpdated: true,
			customCss: [
				'@fontsource-variable/space-grotesk/index.css',
				'@fontsource/space-mono/400.css',
				'@fontsource/space-mono/700.css',
				'./src/styles/global.css',
			],
			plugins: [
				ion({
					footer: {
						text: 'woovi-go-sdk · MIT',
						links: [
							{
								label: 'GitHub',
								link: 'https://github.com/orce-me/woovi-go-sdk',
								newTab: true,
							},
							{
								label: 'Woovi API',
								link: 'https://developers.woovi.com/',
								newTab: true,
							},
						],
					},
				}),
			],
			sidebar: [
				{
					label: '[lucide:home] Home',
					link: '/',
				},
				{
					label: '[lucide:rocket] Guides',
					items: [
						{ label: 'Getting started', slug: 'guides/getting-started' },
						{ label: 'Authentication', slug: 'guides/authentication' },
						{ label: 'Money', slug: 'guides/money' },
						{ label: 'Charges', slug: 'guides/charges' },
						{ label: 'Webhooks', slug: 'guides/webhooks' },
						{ label: 'Pagination', slug: 'guides/pagination' },
					],
				},
				{
					label: '[lucide:book-open] Reference',
					items: [
						{ label: 'Client options', slug: 'reference/client' },
						{ label: 'Resources', slug: 'reference/resources' },
						{ label: 'Errors', slug: 'reference/errors' },
					],
				},
			],
		}),
	],
});
