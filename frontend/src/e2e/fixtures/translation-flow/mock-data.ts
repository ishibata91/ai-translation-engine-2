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
