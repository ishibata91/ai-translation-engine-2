import {defineConfig} from 'astro/config';
import mermaid from 'astro-mermaid';
import starlight from '@astrojs/starlight';

export default defineConfig({
  integrations: [
    mermaid({
      autoTheme: true,
      enableLog: false,
      theme: 'forest',
    }),
    starlight({
      title: 'AI Translation Engine 2 Docs',
      description: 'AI Translation Engine 2 の docs 正本をそのままブラウズする Starlight サイト。',
      credits: true,
      disable404Route: true,
      lastUpdated: true,
    }),
  ],
});
