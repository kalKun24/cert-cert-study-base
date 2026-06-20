import { test, expect } from '@playwright/test';
import { loginAs } from './helpers/auth';
import { USER_IDS, resetSatoTeamOwnerStatus } from './helpers/api';

/**
 * テスト 3: 管理者によるチームオーナー権限設定
 *
 * 3-A: admin でログイン → ユーザー管理 → sato の編集画面 → is_team_owner トグルが OFF
 * 3-B: トグルを ON にして「権限を保存」→ 成功メッセージが出る
 * 3-C: ON の状態で max_teams 入力欄が表示される
 * 3-D: トグルを OFF に戻して保存 → 成功メッセージが出る（OFF に戻す）
 */
test.describe('3: 管理者によるチームオーナー権限設定', () => {
  test.beforeEach(async () => {
    // sato の is_team_owner を false にリセット
    await resetSatoTeamOwnerStatus(false);
  });

  test.afterEach(async () => {
    // テスト後も false に戻す（冪等性を保つ）
    await resetSatoTeamOwnerStatus(false);
  });

  test('3-A: admin でログイン → sato の編集画面で is_team_owner トグルが OFF である', async ({
    page,
  }) => {
    await loginAs(page, 'admin', 'Admin1234!');

    // ユーザー管理ページに移動
    await page.goto(`/admin/users/${USER_IDS.SATO}/edit`);

    const toggle = page.locator('#team-owner-toggle');
    await expect(toggle).toBeVisible();

    // aria-checked が false であることを確認
    await expect(toggle).toHaveAttribute('aria-checked', 'false');
  });

  test('3-B: トグルを ON にして「権限を保存」をクリックすると成功メッセージが表示される', async ({
    page,
  }) => {
    await loginAs(page, 'admin', 'Admin1234!');
    await page.goto(`/admin/users/${USER_IDS.SATO}/edit`);

    const toggle = page.locator('#team-owner-toggle');
    await expect(toggle).toBeVisible();

    // トグルをクリックして ON にする
    await toggle.click();
    await expect(toggle).toHaveAttribute('aria-checked', 'true');

    // max_teams に値を入力（ON になると表示される）
    const maxTeamsInput = page.locator('#max-teams-input');
    await expect(maxTeamsInput).toBeVisible();
    await maxTeamsInput.fill('1');

    // 権限を保存
    await page.getByRole('button', { name: '権限を保存' }).click();

    // 成功メッセージが表示されることを確認
    const successMessage = page.getByRole('status', { name: /権限を更新しました/ });
    await expect(successMessage).toBeVisible();
    await expect(successMessage).toHaveText('権限を更新しました');
  });

  test('3-C: トグルが ON のときに max_teams 入力欄が表示される', async ({ page }) => {
    await loginAs(page, 'admin', 'Admin1234!');
    await page.goto(`/admin/users/${USER_IDS.SATO}/edit`);

    const toggle = page.locator('#team-owner-toggle');
    await expect(toggle).toBeVisible();

    // OFF の状態では max_teams 入力欄が表示されない
    await expect(page.locator('#max-teams-input')).not.toBeVisible();

    // ON にすると表示される
    await toggle.click();
    await expect(page.locator('#max-teams-input')).toBeVisible();
  });

  test('3-D: トグルを OFF に戻して保存すると成功メッセージが表示される', async ({ page }) => {
    // まず ON にしておく
    await resetSatoTeamOwnerStatus(true);

    await loginAs(page, 'admin', 'Admin1234!');
    await page.goto(`/admin/users/${USER_IDS.SATO}/edit`);

    const toggle = page.locator('#team-owner-toggle');
    await expect(toggle).toBeVisible();

    // 現在 ON になっていることを確認
    await expect(toggle).toHaveAttribute('aria-checked', 'true');

    // OFF にする
    await toggle.click();
    await expect(toggle).toHaveAttribute('aria-checked', 'false');

    // 権限を保存
    await page.getByRole('button', { name: '権限を保存' }).click();

    // 成功メッセージが表示されることを確認
    const successMessage = page.getByRole('status', { name: /権限を更新しました/ });
    await expect(successMessage).toBeVisible();
    await expect(successMessage).toHaveText('権限を更新しました');
  });
});
