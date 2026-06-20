import { test, expect } from '@playwright/test';
import { loginAs } from './helpers/auth';
import {
  TEAM_IDS,
  adminLogin,
  loginAs as apiLoginAs,
  listTagsViaApi,
  createTagViaApi,
  deleteTagViaApi,
  cleanupTeamTags,
} from './helpers/api';

/**
 * テスト 8: タグのチームスコープ化
 *
 * 8-A: チームメンバーがタグを作成できる（UIから）
 * 8-B: チームメンバーがタグを削除できる（UIから）
 * 8-C: admin のみタグ名編集ボタンが表示される（メンバーには非表示）
 * 8-D: 別チームのタグはUIに表示されない（チームスコープ分離）
 * 8-E: タグ作成後、問題作成画面でそのタグが選択できる
 */
test.describe('8: タグのチームスコープ化', () => {
  test.afterEach(async () => {
    // 各テスト後にチームAとチームBのタグを全削除してリセット
    await cleanupTeamTags(TEAM_IDS.TEAM_A);
    await cleanupTeamTags(TEAM_IDS.TEAM_B);
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('8-A: チームメンバーがUIからタグを作成できる', async ({ page }) => {
    // tanaka = チームA メンバー（owner）としてログイン
    await loginAs(page, 'tanaka');

    // タグ管理ページへ移動
    await page.goto('/tags');
    await expect(page.locator('.tag-manage-page')).toBeVisible();

    // タグ作成フォームが表示されること（admin でなくてもフォームあり）
    const nameInput = page.locator('#new-tag-name');
    await expect(nameInput).toBeVisible();

    const tagName = `E2E-TagA-${Date.now()}`;
    await nameInput.fill(tagName);
    await page.getByRole('button', { name: '追加' }).click();

    // 一覧に作成したタグが表示される
    await expect(page.locator('.tag-manage-list')).toContainText(tagName);
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('8-B: チームメンバーがUIからタグを削除できる', async ({ page }) => {
    // 事前にAPIでタグを作成しておく
    const tanakaToken = await apiLoginAs('tanaka');
    const tagName = `E2E-DeleteTarget-${Date.now()}`;
    await createTagViaApi(tanakaToken, TEAM_IDS.TEAM_A, tagName);

    await loginAs(page, 'tanaka');
    await page.goto('/tags');
    await expect(page.locator('.tag-manage-list')).toContainText(tagName);

    // 削除ボタンをクリック（確認ダイアログを受諾）
    page.on('dialog', (dialog) => dialog.accept());
    const tagItem = page.locator('.tag-manage-item', { hasText: tagName });
    await tagItem.getByRole('button', { name: '削除' }).click();

    // タグアイテムが一覧から消える（削除成功時はリスト自体もなくなるため tag-manage-item で確認）
    await expect(page.locator('.tag-manage-item', { hasText: tagName })).toHaveCount(0);
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('8-C: admin のみタグ名の編集ボタンが表示され、一般メンバーには表示されない', async ({
    page,
  }) => {
    // 事前にAPIでタグを作成
    const adminToken = await adminLogin();
    const tagName = `E2E-EditCheck-${Date.now()}`;
    await createTagViaApi(adminToken, TEAM_IDS.TEAM_A, tagName);

    // ── admin で確認 ──
    await loginAs(page, 'admin');
    await page.goto('/tags');
    const adminTagItem = page.locator('.tag-manage-item', { hasText: tagName });
    await expect(adminTagItem.getByRole('button', { name: '編集' })).toBeVisible();

    // ── 一般メンバー（tanaka）で確認 ──
    await loginAs(page, 'tanaka');
    await page.goto('/tags');
    const memberTagItem = page.locator('.tag-manage-item', { hasText: tagName });
    await expect(memberTagItem.getByRole('button', { name: '編集' })).not.toBeVisible();
    // 削除ボタンは表示される
    await expect(memberTagItem.getByRole('button', { name: '削除' })).toBeVisible();
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('8-D: チームAのタグはチームBのメンバーのUIに表示されない', async ({ page }) => {
    // チームAに識別しやすいタグを作成
    const tanakaToken = await apiLoginAs('tanaka');
    const teamATagName = `E2E-TeamA-Secret-${Date.now()}`;
    await createTagViaApi(tanakaToken, TEAM_IDS.TEAM_A, teamATagName);

    // チームBに別のタグを作成（チームBのUI表示確認用）
    const suzukiToken = await apiLoginAs('suzuki');
    const teamBTagName = `E2E-TeamB-Tag-${Date.now()}`;
    await createTagViaApi(suzukiToken, TEAM_IDS.TEAM_B, teamBTagName);

    // suzuki（チームBメンバー）でログイン → タグ管理ページ
    await loginAs(page, 'suzuki');
    await page.goto('/tags');

    // チームBのタグは見える
    await expect(page.locator('.tag-manage-page')).toContainText(teamBTagName);

    // チームAのタグは見えない
    await expect(page.locator('.tag-manage-page')).not.toContainText(teamATagName);
  });

  test('8-D (API): チームAのタグはチームBメンバーから403が返る', async () => {
    // チームAにタグを作成
    const tanakaToken = await apiLoginAs('tanaka');
    await createTagViaApi(tanakaToken, TEAM_IDS.TEAM_A, `E2E-API-Scope-${Date.now()}`);

    // suzuki（チームBのみ所属）でチームAのタグ一覧を取得 → 403
    const suzukiToken = await apiLoginAs('suzuki');
    const res = await fetch(`http://localhost:8080/api/v1/teams/${TEAM_IDS.TEAM_A}/tags`, {
      headers: { Authorization: `Bearer ${suzukiToken}` },
    });
    expect(res.status).toBe(403);
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('8-E: タグ作成後、問題作成画面でそのタグが選択できる', async ({ page }) => {
    // チームAにタグを作成しておく
    const tanakaToken = await apiLoginAs('tanaka');
    const tagName = `E2E-QuestionTag-${Date.now()}`;
    await createTagViaApi(tanakaToken, TEAM_IDS.TEAM_A, tagName);

    // tanaka でログイン → 問題作成ページへ
    await loginAs(page, 'tanaka');
    await page.goto('/questions/new');

    // タグ選択チェックボックス欄にタグが表示される
    const tagCheckbox = page.locator('.tag-checkbox-label', { hasText: tagName });
    await expect(tagCheckbox).toBeVisible();

    // チェックできる
    await tagCheckbox.locator('input[type="checkbox"]').check();
    await expect(tagCheckbox.locator('input[type="checkbox"]')).toBeChecked();
  });

  // ───────────────────────────────────────────────────────────────────────────

  test('8-F: admin が別チームのタグをAPIで削除できる', async () => {
    // チームAにタグを作成（tanaka で）
    const tanakaToken = await apiLoginAs('tanaka');
    const tagName = `E2E-AdminDelete-${Date.now()}`;
    const tagId = await createTagViaApi(tanakaToken, TEAM_IDS.TEAM_A, tagName);

    // admin でチームAのタグを削除
    const adminToken = await adminLogin();
    await deleteTagViaApi(adminToken, TEAM_IDS.TEAM_A, tagId);

    // 削除後は一覧に存在しない
    const tags = await listTagsViaApi(adminToken, TEAM_IDS.TEAM_A);
    expect(tags.some((t) => t.id === tagId)).toBe(false);
  });
});
