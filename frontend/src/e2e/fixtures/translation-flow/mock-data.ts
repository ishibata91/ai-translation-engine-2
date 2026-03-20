const FILE_A_ID = 101;
const FILE_B_ID = 102;

const buildDialogueRows = (count: number) =>
  Array.from({length: count}, (_, index) => {
    const rowNumber = index + 1;
    const padded = String(rowNumber).padStart(3, '0');
    return {
      id: `dialogue_response:${rowNumber}`,
      section: 'dialogue_response',
      record_type: 'INFO',
      editor_id: `EDID_${padded}`,
      source_text: `Dialogue line ${padded}`,
    };
  });

const FILE_A_ROWS = buildDialogueRows(53);
const FILE_B_ROWS = [
  {
    id: 'quest_objective:1',
    section: 'quest_objective',
    record_type: 'QUST',
    editor_id: 'Q_B_01',
    source_text: 'Quest Objective B-01',
  },
  {
    id: 'item_name:2',
    section: 'item_name',
    record_type: 'BOOK',
    editor_id: 'BOOK_B_02',
    source_text: 'Book Title B-02',
  },
  {
    id: 'npc_name:3',
    section: 'npc_name',
    record_type: 'NPC_',
    editor_id: 'NPC_B_03',
    source_text: 'NPC Name B-03',
  },
];

export const TRANSLATION_FLOW_TASK_ID = 'translation-flow-e2e-task';

export const TRANSLATION_FLOW_SELECTED_FILES = [
  'C:/fixtures/translation-flow/Skyrim.esm.extract.json',
  'C:/fixtures/translation-flow/Update.esm.extract.json',
] as const;

export const TRANSLATION_FLOW_FILE_PAYLOADS = {
  [TRANSLATION_FLOW_SELECTED_FILES[0]]: {
    file_id: FILE_A_ID,
    file_path: TRANSLATION_FLOW_SELECTED_FILES[0],
    file_name: 'Skyrim.esm.extract.json',
    parse_status: 'loaded',
    rows: FILE_A_ROWS,
  },
  [TRANSLATION_FLOW_SELECTED_FILES[1]]: {
    file_id: FILE_B_ID,
    file_path: TRANSLATION_FLOW_SELECTED_FILES[1],
    file_name: 'Update.esm.extract.json',
    parse_status: 'loaded',
    rows: FILE_B_ROWS,
  },
} as const;

export const TRANSLATION_FLOW_EXPECTED_FILE_NAMES = ['Skyrim.esm.extract.json', 'Update.esm.extract.json'] as const;

export const TRANSLATION_FLOW_PAGE_TWO_MARKER = 'Dialogue line 051';
export const TRANSLATION_FLOW_SECOND_FILE_MARKER = 'Quest Objective B-01';

export const TRANSLATION_FLOW_TERMINOLOGY_SUMMARY = {
  task_id: TRANSLATION_FLOW_TASK_ID,
  status: 'pending',
  saved_count: 0,
  failed_count: 0,
  progress_mode: 'hidden',
  progress_current: 0,
  progress_total: 0,
  progress_message: '',
} as const;

export const TRANSLATION_FLOW_TERMINOLOGY_COMPLETED_SUMMARY = {
  task_id: TRANSLATION_FLOW_TASK_ID,
  status: 'completed',
  saved_count: 8,
  failed_count: 0,
  progress_mode: 'hidden',
  progress_current: 8,
  progress_total: 8,
  progress_message: '',
} as const;

export const TRANSLATION_FLOW_TERMINOLOGY_TARGET_PAGE = {
  task_id: TRANSLATION_FLOW_TASK_ID,
  page: 1,
  page_size: 50,
  total_rows: 8,
  rows: [
    {
      id: 'npc-full-01',
      record_type: 'NPC_:FULL',
      editor_id: 'NPC_B_03',
      source_text: 'NPC Name B-03',
      translated_text: '',
      translation_state: 'missing',
      variant: 'full',
      source_file: 'Update.esm.extract.json',
    },
    {
      id: 'npc-short-01',
      record_type: 'NPC_:SHRT',
      editor_id: 'NPC_B_03',
      source_text: 'NPC Name B-03',
      translated_text: '',
      translation_state: 'missing',
      variant: 'short',
      source_file: 'Update.esm.extract.json',
    },
    {
      id: 'quest-01',
      record_type: 'QUST',
      editor_id: 'Q_B_01',
      source_text: 'Quest Objective B-01',
      translated_text: '',
      translation_state: 'missing',
      variant: 'single',
      source_file: 'Update.esm.extract.json',
    },
  ],
} as const;

export const TRANSLATION_FLOW_TERMINOLOGY_COMPLETED_TARGET_PAGE = {
  ...TRANSLATION_FLOW_TERMINOLOGY_TARGET_PAGE,
  rows: [
    {
      ...TRANSLATION_FLOW_TERMINOLOGY_TARGET_PAGE.rows[0],
      translated_text: 'NPC 名 B-03',
      translation_state: 'translated',
    },
    {
      ...TRANSLATION_FLOW_TERMINOLOGY_TARGET_PAGE.rows[1],
      translated_text: 'NPC 名',
      translation_state: 'translated',
    },
    {
      ...TRANSLATION_FLOW_TERMINOLOGY_TARGET_PAGE.rows[2],
      translated_text: 'クエスト目標 B-01',
      translation_state: 'translated',
    },
  ],
} as const;

export const TRANSLATION_FLOW_TERMINOLOGY_LLM_CONFIG = {
  provider: 'lmstudio',
  model: '',
  endpoint: 'http://localhost:1234',
  api_key: '',
  temperature: '0.2',
  context_length: '0',
  sync_concurrency: '2',
  bulk_strategy: 'sync',
} as const;

export const TRANSLATION_FLOW_TERMINOLOGY_PROMPT_CONFIG = {
  user_prompt: 'Translate the provided term.',
  system_prompt: 'You are a translator for a Skyrim mod.',
} as const;
