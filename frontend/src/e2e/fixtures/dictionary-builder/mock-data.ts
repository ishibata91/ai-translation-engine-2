type DictionaryMockSource = {
  entry_count: number;
  error_message: string | null;
  file_name: string;
  file_path: string;
  file_size_bytes: number;
  format: string;
  id: number;
  imported_at: string;
  status: string;
};

type DictionaryMockEntry = {
  dest_text: string;
  edid: string;
  id: number;
  record_type: string;
  source_id: number;
  source_name: string;
  source_text: string;
};

export const DICTIONARY_BUILDER_SOURCES: DictionaryMockSource[] = [
  {
    id: 101,
    file_name: 'Skyrim.esm.xml',
    format: 'sstxml',
    entry_count: 2,
    status: 'COMPLETED',
    imported_at: '2026-03-10T09:30:00Z',
    file_path: 'frontend/src/e2e/fixtures/dictionary-builder/source-skyrim.esm.xml',
    file_size_bytes: 4096,
    error_message: null,
  },
  {
    id: 102,
    file_name: 'Dawnguard.esm.xml',
    format: 'sstxml',
    entry_count: 1,
    status: 'COMPLETED',
    imported_at: '2026-03-10T10:00:00Z',
    file_path: 'frontend/src/e2e/fixtures/dictionary-builder/source-dawnguard.esm.xml',
    file_size_bytes: 3072,
    error_message: null,
  },
];

export const DICTIONARY_BUILDER_ENTRIES_BY_SOURCE_ID: Record<number, DictionaryMockEntry[]> = {
  101: [
    {
      id: 1,
      source_id: 101,
      source_name: 'Skyrim.esm.xml',
      edid: 'MAG_Fireball',
      source_text: 'Cast a blazing fireball.',
      dest_text: '灼熱のファイアボールを放つ。',
      record_type: 'INFO',
    },
    {
      id: 2,
      source_id: 101,
      source_name: 'Skyrim.esm.xml',
      edid: 'NPC_GuardWarning',
      source_text: 'Halt, you have violated the law.',
      dest_text: '待て、法を犯したな。',
      record_type: 'DIAL',
    },
  ],
  102: [
    {
      id: 3,
      source_id: 102,
      source_name: 'Dawnguard.esm.xml',
      edid: 'DLG_DragonHunter',
      source_text: 'The dragon circles above the fort.',
      dest_text: 'ドラゴンが砦の上空を旋回している。',
      record_type: 'INFO',
    },
  ],
};

export const DICTIONARY_BUILDER_REQUIRED_QUERY = 'dragon';

export const DICTIONARY_BUILDER_REQUIRED_SOURCE_FILES = DICTIONARY_BUILDER_SOURCES.map(
  (source) => source.file_name,
);

export const DICTIONARY_BUILDER_REQUIRED_ENTRY_MARKERS = ['MAG_Fireball', 'NPC_GuardWarning'];

export const DICTIONARY_BUILDER_REQUIRED_CROSS_SEARCH_MARKERS = ['DLG_DragonHunter'];
