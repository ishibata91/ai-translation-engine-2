import {test} from './fixtures/app.fixture';
import {DASHBOARD_DELETE_TASK_ID, DASHBOARD_DELETE_TASK_NAME} from './fixtures/dashboard/mock-data';

test('Dashboard の必須シナリオ: 確認 modal 経由で task を削除できる', async ({app}) => {
  await app.dashboard.open();
  await app.dashboard.expectTaskVisible(DASHBOARD_DELETE_TASK_NAME);
  await app.dashboard.openDeleteModal(DASHBOARD_DELETE_TASK_ID);
  await app.dashboard.confirmDelete();
  await app.dashboard.expectTaskNotVisible(DASHBOARD_DELETE_TASK_NAME);
});
