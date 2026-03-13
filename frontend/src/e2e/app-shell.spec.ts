import {test} from './fixtures/app.fixture';

test('アプリ起動時に主要レイアウトが表示される', async ({ app }) => {
  await app.openDashboard();
  await app.expectShellVisible();
});

test('ダッシュボードが初期表示される', async ({ app }) => {
  await app.openDashboard();
  await app.expectDashboardVisible();
});

test('サイドバーから辞書構築へ遷移できる', async ({ app }) => {
  await app.openDashboard();
  await app.openDictionaryBySidebar();
  await app.expectDictionaryVisible();
});

test('サイドバーからマスターペルソナ構築へ遷移できる', async ({ app }) => {
  await app.openDashboard();
  await app.openMasterPersonaBySidebar();
  await app.expectMasterPersonaVisible();
});
