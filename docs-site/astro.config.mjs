import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  base: process.env.SITE_BASE,
  outDir: '../docs/roadmap',
  integrations: [
    starlight({
      title: 'Cassor',
      description: 'Developer roadmap and project documentation.',
      sidebar: [
        { label: 'Roadmap', link: '/' },
        {
          label: 'Feature details',
          items: [{ autogenerate: { directory: 'roadmap' } }],
        },
        {
          label: 'Documentation',
          items: [{ autogenerate: { directory: 'guides' } }],
        },
      ],
    }),
  ],
});
