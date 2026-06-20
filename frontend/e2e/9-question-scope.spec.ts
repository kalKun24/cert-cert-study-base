import { test, expect } from '@playwright/test';
import { loginAs } from './helpers/auth';
import {
  TEAM_IDS,
  loginAs as apiLoginAs,
  createQuestionViaApi,
  deleteQuestionViaApi,
  listQuestionsViaApi,
  cleanupTeamQuestions,
} from './helpers/api';

/**
 * テスト 9: 問題のチームスコープ化
 *
 * 9-A: チームAメンバーがUIから問題を作成できる
 * 9-B: チームAの問題がチームBのメンバーのUIに表示されない
 * 9-C (API): チームBメンバーがチームAの問題一覧にアクセスすると403
 * 9-D (API): チームAに属さないユーザー（yamada→TeamB）がチームBにアクセスすると403
 * 9-E: チームAメンバーが問題詳細をUIで閲覧できる
 * 9-F: チームAメンバーがUIから問題を削除できる
 */
test.describe('9: 問題のチームスコープ化', () => {
  test.afterEach(async () => {
    await cleanupTeamQuestions(TEAM_IDS.TEAM_A);
    await cleanupTeamQuestions(TEAM_IDS.TEAM_B);
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('9-A: チームAメンバーがUIから問題を作成できる', async ({ page }) => {
    await loginAs(page, 'tanaka');

    await page.goto('/questions/new');
    await expect(page.locator('.question-form-page')).toBeVisible();

    const title = `E2E-問題作成-${Date.now()}`;

    // タイトル入力
    await page.locator('#question-title').fill(title);

    // 問題文（MDEditor の textarea に入力）
    const bodyTab = page.getByRole('tab', { name: '問題文' });
    await bodyTab.click();
    // MDEditor の textarea
    const textarea = page.locator('.w-md-editor-text-input').first();
    await textarea.fill('## 問題\nE2Eテスト用の問題文です。');

    // 送信
    await page.getByRole('button', { name: '作成' }).click();

    // 詳細ページに遷移すること
    await page.waitForURL(/\/questions\/[^/]+$/);
    await expect(page.locator('.question-detail-title')).toContainText(title);
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('9-B: チームAの問題はチームBメンバーのUIに表示されない', async ({ page }) => {
    // チームAに問題を作成
    const tanakaToken = await apiLoginAs('tanaka');
    const title = `E2E-TeamA専用-${Date.now()}`;
    await createQuestionViaApi(tanakaToken, TEAM_IDS.TEAM_A, title);

    // suzuki（チームBのみ所属）でログイン → 問題一覧
    await loginAs(page, 'suzuki');
    await page.goto('/questions');
    await expect(page.locator('.question-list-page')).toBeVisible();

    // チームAの問題タイトルは表示されない
    await expect(page.locator('.question-list-page')).not.toContainText(title);
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('9-C (API): チームBメンバーがチームAの問題一覧にアクセスすると403', async () => {
    const suzukiToken = await apiLoginAs('suzuki');
    const res = await listQuestionsViaApi(suzukiToken, TEAM_IDS.TEAM_A);
    expect(res.status).toBe(403);
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('9-D (API): チームAメンバーがチームBの問題一覧にアクセスすると403', async () => {
    // tanaka はチームAのメンバーでチームBには所属していない
    const tanakaToken = await apiLoginAs('tanaka');
    const res = await listQuestionsViaApi(tanakaToken, TEAM_IDS.TEAM_B);
    expect(res.status).toBe(403);
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('9-D-admin (API): adminもチームメンバーでなければチームBの問題一覧で403', async () => {
    // チームBにチームA問題を作成（suzuki で）
    const suzukiToken = await apiLoginAs('suzuki');
    await createQuestionViaApi(suzukiToken, TEAM_IDS.TEAM_B, `E2E-TeamB-${Date.now()}`);

    // admin でチームBの問題一覧取得 → admin はチームBのメンバーでないため403
    const { adminLogin } = await import('./helpers/api');
    const adminToken = await adminLogin();
    const res = await listQuestionsViaApi(adminToken, TEAM_IDS.TEAM_B);
    // admin が Team B のメンバーでなければ 403
    expect(res.status).toBe(403);
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('9-E: チームAメンバーが問題詳細をUIで閲覧できる', async ({ page }) => {
    // APIで問題を作成
    const tanakaToken = await apiLoginAs('tanaka');
    const title = `E2E-詳細閲覧-${Date.now()}`;
    const questionId = await createQuestionViaApi(tanakaToken, TEAM_IDS.TEAM_A, title);

    await loginAs(page, 'tanaka');
    await page.goto(`/questions/${questionId}`);

    // 詳細ページが表示される
    await expect(page.locator('.question-detail-title')).toContainText(title);
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('9-F: チームAメンバーがUIから問題を削除できる', async ({ page }) => {
    // APIで問題を作成
    const tanakaToken = await apiLoginAs('tanaka');
    const title = `E2E-削除テスト-${Date.now()}`;
    const questionId = await createQuestionViaApi(tanakaToken, TEAM_IDS.TEAM_A, title);

    await loginAs(page, 'tanaka');
    await page.goto(`/questions/${questionId}`);
    await expect(page.locator('.question-detail-title')).toContainText(title);

    // 削除ボタンをクリック（確認ダイアログを受諾）
    page.on('dialog', (dialog) => dialog.accept());
    await page.getByRole('button', { name: '削除' }).click();

    // 問題一覧ページに戻ること
    await page.waitForURL('/questions');
    await expect(page.locator('.question-list-page')).toBeVisible();

    // 削除した問題がAPIでも取得できないこと（404）
    const res = await fetch(
      `http://localhost:8080/api/v1/teams/${TEAM_IDS.TEAM_A}/questions/${questionId}`,
      { headers: { Authorization: `Bearer ${tanakaToken}` } },
    );
    expect(res.status).toBe(404);
  });
});
