import { test, expect } from '@playwright/test';
import { loginAs } from './helpers/auth';
import { TEAM_IDS } from './helpers/api';

/**
 * テスト 7: 権限エラー
 *
 * 7-A: sato でログイン → /teams/new に直アクセス → 403 画面またはリダイレクトされる
 * 7-B: suzuki でログイン → チームを作成しようとすると max_teams=1 を超えてエラーになる
 *       （suzuki はすでにチームBを持っている）
 */
test.describe('7: 権限エラー', () => {
  test('7-A: sato（is_team_owner=false）で /teams/new に直アクセスするとリダイレクトされる', async ({
    page,
  }) => {
    await loginAs(page, 'sato');

    // /teams/new に直アクセス
    await page.goto('/teams/new');

    // PrivateRoute が requiredRoles=['admin', 'teamowner'] で守っているため
    // sato（role=user, is_team_owner=false）はアクセスできず / にリダイレクトされる
    await page.waitForURL(
      (url) =>
        url.pathname !== '/teams/new',
      { timeout: 10000 },
    );

    const pathname = new URL(page.url()).pathname;
    // /teams/new 以外の画面にいることを確認
    expect(pathname).not.toBe('/teams/new');
    // ログインページやホームにリダイレクトされている
    expect(['/', '/login', '/invitations', '/no-team']).toContain(pathname);
  });

  test('7-B: suzuki がチーム作成フォームを送信すると max_teams 超過エラーが表示される', async ({
    page,
  }) => {
    // suzuki は is_team_owner=true, max_teams=1 で、すでにチームBを所有している
    await loginAs(page, 'suzuki');

    // チーム作成ページに移動（suzuki は is_team_owner=true なのでアクセス可能）
    await page.goto('/teams/new');

    // チーム作成フォームが表示されることを確認
    const nameInput = page.getByLabel('チーム名');
    await expect(nameInput).toBeVisible();

    // フォームに入力して送信
    await nameInput.fill('テスト用チーム');

    const descInput = page.getByLabel('説明');
    if (await descInput.isVisible()) {
      await descInput.fill('テスト用の説明');
    }

    // 作成ボタンをクリック
    const submitButton = page.getByRole('button', { name: '作成' });
    await submitButton.click();

    // max_teams=1 を超えているためエラーが表示されることを確認
    // エラーメッセージまたはアラートが表示される
    const errorAlert = page.locator('[role="alert"]');
    await expect(errorAlert).toBeVisible({ timeout: 10000 });
  });
});
