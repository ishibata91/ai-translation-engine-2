type PersonaNpcRecord = {
  dialogue_count: number;
  generation_request: string;
  npc_name: string;
  persona_id: number;
  persona_text: string;
  race: string;
  sex: string;
  source_plugin: string;
  speaker_id: string;
  status: 'draft' | 'generated';
  updated_at: string;
  voice_type: string;
};

type PersonaDialogueRecord = {
  editor_id: string;
  record_type: string;
  source_text: string;
};

type ModelCatalogEntry = {
  capability: {
    supports_batch: boolean;
  };
  display_name: string;
  id: string;
};

export const MASTER_PERSONA_REQUIRED_NPCS: PersonaNpcRecord[] = [
  {
    persona_id: 2001,
    speaker_id: '00013BA3',
    source_plugin: 'Skyrim.esm',
    npc_name: 'AelaTheHuntress',
    race: 'NordRace',
    sex: 'Female',
    voice_type: 'FemaleNord',
    dialogue_count: 2,
    status: 'generated',
    updated_at: '2026-03-10T12:00:00Z',
    persona_text: '狩人として誇り高く、簡潔で力強い語り口。',
    generation_request: 'TL: |Personality Traits: stoic hunter|',
  },
  {
    persona_id: 2002,
    speaker_id: '0001414D',
    source_plugin: 'Skyrim.esm',
    npc_name: 'BalgruufTheGreater',
    race: 'NordRace',
    sex: 'Male',
    voice_type: 'MaleNord',
    dialogue_count: 1,
    status: 'draft',
    updated_at: '2026-03-11T01:20:00Z',
    persona_text: '統治者らしい落ち着きと責任感がある。',
    generation_request: 'TL: |Personality Traits: pragmatic ruler|',
  },
];

export const MASTER_PERSONA_DIALOGUES_BY_PERSONA_ID: Record<number, PersonaDialogueRecord[]> = {
  2001: [
    {
      editor_id: 'AELA_GREET_01',
      record_type: 'INFO',
      source_text: 'The blood of your foes is all the glory you need.',
    },
    {
      editor_id: 'AELA_MISC_02',
      record_type: 'INFO',
      source_text: 'Kodlak taught us to hunt with honor.',
    },
  ],
  2002: [
    {
      editor_id: 'BALGRUUF_WARN_01',
      record_type: 'INFO',
      source_text: 'Whiterun must stand strong against any threat.',
    },
  ],
};

export const MASTER_PERSONA_PROMPT_CONFIG = {
  user_prompt: 'NPC の口調、一人称、二人称、命令調の有無を重点分析する。',
  system_prompt: 'You are a character persona analyzer for RPG dialogue.',
};

export const MASTER_PERSONA_LLM_ROOT_CONFIG: Record<string, string> = {
  selected_provider: 'lmstudio',
  'sync_concurrency.lmstudio': '4',
};

export const MASTER_PERSONA_LLM_CONFIG_BY_NAMESPACE: Record<string, Record<string, string>> = {
  'master_persona.llm.lmstudio': {
    model: 'local-model',
    endpoint: 'http://localhost:1234',
    api_key: '',
    temperature: '0.30',
    context_length: '8192',
  },
  'master_persona.llm.gemini': {
    model: 'gemini-2.0-flash',
    endpoint: 'https://generativelanguage.googleapis.com',
    api_key: 'gm-e2e-key',
    temperature: '0.50',
    context_length: '16384',
    bulk_strategy: 'batch',
  },
  'master_persona.llm.xai': {
    model: 'grok-3-mini',
    endpoint: 'https://api.x.ai/v1',
    api_key: 'xai-e2e-key',
    temperature: '0.40',
    context_length: '16384',
    bulk_strategy: 'batch',
  },
};

export const MASTER_PERSONA_MODEL_CATALOG_BY_PROVIDER: Record<string, ModelCatalogEntry[]> = {
  lmstudio: [
    {id: 'local-model', display_name: 'local-model', capability: {supports_batch: false}},
    {id: 'local-model-2', display_name: 'local-model-2', capability: {supports_batch: false}},
  ],
  gemini: [
    {id: 'gemini-2.0-flash', display_name: 'gemini-2.0-flash', capability: {supports_batch: true}},
    {id: 'gemini-2.0-pro', display_name: 'gemini-2.0-pro', capability: {supports_batch: true}},
  ],
  xai: [
    {id: 'grok-3-mini', display_name: 'grok-3-mini', capability: {supports_batch: true}},
    {id: 'grok-3', display_name: 'grok-3', capability: {supports_batch: true}},
  ],
};

export const MASTER_PERSONA_SELECTED_JSON_PATH = 'frontend/tests/e2e/fixtures/master-persona/Dawnguard.esm_Export.json';
export const MASTER_PERSONA_STARTED_TASK_ID = 'persona-task-required-e2e';

export const MASTER_PERSONA_REQUIRED_NAMES = MASTER_PERSONA_REQUIRED_NPCS.map((entry) => entry.npc_name);
export const MASTER_PERSONA_REQUIRED_PROMPT_TEXT = MASTER_PERSONA_PROMPT_CONFIG.user_prompt;
