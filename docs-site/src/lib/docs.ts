import type {CollectionEntry} from 'astro:content';

export type DocsEntry = CollectionEntry<'docs'>;
export type ChangeEntry = CollectionEntry<'changes'>;
export type SkillEntry = CollectionEntry<'skills'>;
export type SiteEntry = DocsEntry | ChangeEntry | SkillEntry;

type SidebarLink = {
  label: string;
  link: string;
};

type SidebarGroup = {
  label: string;
  items: SidebarItem[];
  collapsed?: boolean;
};

export type SidebarItem = SidebarLink | SidebarGroup;

type TreeNode = {
  label: string;
  link?: string;
  children: Map<string, TreeNode>;
};

const ZONE_LABELS: Record<string, string> = {
  governance: 'Governance',
  frontend: 'Frontend',
  controller: 'Controller',
  workflow: 'Workflow',
  slice: 'Slice',
  runtime: 'Runtime',
  artifact: 'Artifact',
  gateway: 'Gateway',
  foundation: 'Foundation',
  changes: 'Changes',
  skills: 'Skills',
  references: 'References',
};

function slugifySegment(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[_\s]+/g, '-')
    .replace(/[^a-z0-9-]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

function extractTitleFromBody(body: string): string | undefined {
  const heading = body.match(/^#\s+(.+)$/m);
  return heading?.[1]?.trim();
}

function extractDescriptionFromBody(body: string): string | undefined {
  const lines = body.split(/\r?\n/);
  let afterHeading = false;

  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (!afterHeading) {
      if (line.startsWith('# ')) {
        afterHeading = true;
      }
      continue;
    }

    if (!line || line.startsWith('#') || line.startsWith('```')) {
      continue;
    }

    return line;
  }

  return undefined;
}

function prettifyName(value: string): string {
  return value
    .replace(/[_-]+/g, ' ')
    .replace(/\b\w/g, (char) => char.toUpperCase());
}

function getRelativeParts(entry: SiteEntry): string[] {
  return entry.id.split('/').filter(Boolean);
}

function isDirectoryIndexName(value: string): boolean {
  const normalized = value.toLowerCase();
  return normalized === 'index' || normalized === 'spec' || normalized === 'skill';
}

export function getRouteSegments(entry: SiteEntry): string[] {
  const parts = getRelativeParts(entry);
  const last = parts.at(-1);

  if (!last) {
    return [];
  }

  if (entry.collection === 'changes') {
    if (parts.length === 1 && last === 'index') {
      return ['changes'];
    }

    if (isDirectoryIndexName(last)) {
      return ['changes', ...parts.slice(0, -1).map(slugifySegment)];
    }

    return ['changes', ...parts.slice(0, -1).map(slugifySegment), slugifySegment(last)];
  }

  if (entry.collection === 'skills') {
    if (parts.length === 1 && last === 'index') {
      return ['skills'];
    }

    if (isDirectoryIndexName(last)) {
      return ['skills', ...parts.slice(0, -1).map(slugifySegment)];
    }

    return ['skills', ...parts.slice(0, -1).map(slugifySegment), slugifySegment(last)];
  }

  if (parts.length === 1) {
    if (last === 'index') {
      return [];
    }
    return ['references', slugifySegment(last)];
  }

  if (isDirectoryIndexName(last)) {
    return parts.slice(0, -1).map(slugifySegment);
  }

  return parts.slice(0, -1).map(slugifySegment).concat(slugifySegment(last));
}

export function getDocUrl(entry: SiteEntry): string {
  const segments = getRouteSegments(entry);
  return segments.length === 0 ? '/' : `/${segments.join('/')}/`;
}

export function getDocTitle(entry: SiteEntry): string {
  return entry.data.title ?? extractTitleFromBody(entry.body) ?? prettifyName(getRelativeParts(entry).at(-1) ?? entry.id);
}

export function getDocDescription(entry: SiteEntry): string | undefined {
  return entry.data.description ?? extractDescriptionFromBody(entry.body);
}

function getZoneKey(entry: SiteEntry): string {
  const [first] = getRouteSegments(entry);
  return first ?? 'references';
}

function isSidebarGroup(item: SidebarItem): item is SidebarGroup {
  return 'items' in item;
}

function createTreeNode(label: string): TreeNode {
  return {
    label,
    children: new Map<string, TreeNode>(),
  };
}

function getGroupLabel(segment: string, depth: number): string {
  if (depth === 0) {
    return ZONE_LABELS[segment] ?? prettifyName(segment);
  }

  return prettifyName(segment);
}

function ensureChildNode(parent: TreeNode, key: string, label: string): TreeNode {
  const existing = parent.children.get(key);
  if (existing) {
    return existing;
  }

  const node = createTreeNode(label);
  parent.children.set(key, node);
  return node;
}

function addEntryToTree(root: TreeNode, entry: SiteEntry): void {
  const segments = getRouteSegments(entry);
  const title = getDocTitle(entry);
  const url = getDocUrl(entry);

  if (segments.length === 0) {
    root.link = url;
    root.label = title;
    return;
  }

  let cursor = root;

  for (let index = 0; index < segments.length; index += 1) {
    const segment = segments[index];
    const isLeaf = index === segments.length - 1;
    const child = ensureChildNode(cursor, segment, isLeaf ? title : getGroupLabel(segment, index));

    if (isLeaf) {
      child.label = title;
      child.link = url;
    }

    cursor = child;
  }
}

function treeNodeToSidebarItem(node: TreeNode): SidebarItem {
  const childItems = Array.from(node.children.values())
    .sort((left, right) => left.label.localeCompare(right.label))
    .map(treeNodeToSidebarItem);

  if (childItems.length === 0 && node.link) {
    return {
      label: node.label,
      link: node.link,
    };
  }

  const items = childItems;

  if (node.link && !items.some((item) => !isSidebarGroup(item) && item.link === node.link)) {
    items.unshift({
      label: 'Overview',
      link: node.link,
    });
  }

  return {
    label: node.label,
    collapsed: false,
    items,
  };
}

export function buildSidebar(entries: SiteEntry[]): SidebarItem[] {
  const orderedZones = [
    'governance',
    'frontend',
    'controller',
    'workflow',
    'slice',
    'runtime',
    'artifact',
    'gateway',
    'foundation',
    'changes',
    'skills',
    'references',
  ];

  const root = createTreeNode('Spec Home');

  for (const entry of entries) {
    addEntryToTree(root, entry);
  }

  const sidebar: SidebarItem[] = [];

  if (root.link) {
    sidebar.push({ label: root.label, link: root.link });
  }

  for (const zone of orderedZones) {
    const node = root.children.get(zone);
    if (!node) {
      continue;
    }

    node.label = ZONE_LABELS[zone] ?? prettifyName(zone);
    sidebar.push(treeNodeToSidebarItem(node));
  }

  return sidebar;
}
