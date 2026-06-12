import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';

// Served from the custom domain deepseek-orca.io at the site root.
export default defineConfig({
  site: 'https://deepseek-orca.io',
  build: { assets: 'static' },
  integrations: [sitemap()],
});
