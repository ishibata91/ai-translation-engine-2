import {test} from './fixtures/app.fixture';
import {
    MASTER_PERSONA_REQUIRED_NAMES,
    MASTER_PERSONA_REQUIRED_NPCS,
    MASTER_PERSONA_REQUIRED_PROMPT_TEXT,
    MASTER_PERSONA_SELECTED_JSON_PATH,
} from './fixtures/master-persona/mock-data';

const PRIMARY_PERSONA = MASTER_PERSONA_REQUIRED_NPCS[0];

test('MasterPersona の必須シナリオ: 初期表示', async ({app}) => {
  await app.masterPersona.open();
  await app.masterPersona.expectVisible();
  await app.masterPersona.expectNpcList(MASTER_PERSONA_REQUIRED_NAMES);
});

test('MasterPersona の必須シナリオ: NPC 選択で詳細確認', async ({app}) => {
  await app.masterPersona.open();
  await app.masterPersona.selectNpc(PRIMARY_PERSONA.npc_name);
  await app.masterPersona.expectPersonaDetail({
    formId: PRIMARY_PERSONA.speaker_id,
    name: PRIMARY_PERSONA.npc_name,
  });
});

test('MasterPersona の必須シナリオ: PromptSettingCard の表示と編集境界', async ({app}) => {
  await app.masterPersona.open();
  await app.masterPersona.expectPromptCards();
  await app.masterPersona.expectSystemPromptReadonly();
  await app.masterPersona.editUserPrompt(`${MASTER_PERSONA_REQUIRED_PROMPT_TEXT} / edited by e2e`);
  await app.masterPersona.expectUserPromptValue(`${MASTER_PERSONA_REQUIRED_PROMPT_TEXT} / edited by e2e`);
  await app.masterPersona.expectSystemPromptReadonly();
});

test('MasterPersona の必須シナリオ: ModelSettings の主要操作', async ({app}) => {
  await app.masterPersona.open();
  await app.masterPersona.changeProvider('openai');
  await app.masterPersona.expectModelSettingsValue({
    model: 'gpt-4o-mini',
    provider: 'openai',
  });
  await app.masterPersona.changeTemperature('0.72');
  await app.masterPersona.expectModelSettingsValue({temperature: '0.72'});
});

test('MasterPersona の必須シナリオ: JSON 選択からタスク開始', async ({app}) => {
  await app.masterPersona.open();
  await app.masterPersona.chooseJson();
  await app.masterPersona.expectStartReady(MASTER_PERSONA_SELECTED_JSON_PATH);
  await app.masterPersona.startTask();
  await app.masterPersona.expectTaskStarted();
});
