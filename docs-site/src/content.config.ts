import {defineCollection, z} from 'astro:content';
import {glob} from 'astro/loaders';

const docs = defineCollection({
  loader: glob({
    base: '../docs',
    pattern: '**/*.md',
    generateId: ({ entry }) => entry.replace(/\\/g, '/').replace(/\.md$/i, ''),
  }),
  schema: z.object({
    title: z.string().optional(),
    description: z.string().optional(),
  }).passthrough(),
});

const changes = defineCollection({
  loader: glob({
    base: '../changes',
    pattern: '**/*.md',
    generateId: ({ entry }) => entry.replace(/\\/g, '/').replace(/\.md$/i, ''),
  }),
  schema: z.object({
    title: z.string().optional(),
    description: z.string().optional(),
  }).passthrough(),
});

const skills = defineCollection({
  loader: glob({
    base: '../skills',
    pattern: '**/*.md',
    generateId: ({ entry }) => entry.replace(/\\/g, '/').replace(/\.md$/i, ''),
  }),
  schema: z.object({
    title: z.string().optional(),
    description: z.string().optional(),
  }).passthrough(),
});

export const collections = { docs, changes, skills };
