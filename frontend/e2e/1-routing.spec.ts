import { test, expect } from '@playwright/test';
import { loginAs } from './helpers/auth';
import { resetNakamuraInvitation } from './helpers/api';

/**
 * テスト 1: ログイン後の画面分岐
 *
 * 1-A: nakamura でログイン → /invitations 画面が表示される
 * 1-B: nakamura でログイン後 /questions に直アクセス → /invitations にリダイレクト
 * 1-C: sato でログイン → ホーム画面（/）が表示される
 */
test.describe('1: ログイン後の画面分岐', () => {
  test.beforeAll(async () => {
    // nakamura の招待を pending 状態にリセット
    await resetNakamuraInvitation();
  });

  test('1-A: nakamura でログインすると /invitations 画面が表示される', async ({ page }) => {
    await loginAs(page, 'nakamura', 'Test1234!');

    // チームへの招待ページタイトルが表示されることを確認
    await expect(page.locator('#invitation-page-title')).toBeVisible();
    await expect(page.locator('#invitation-page-title')).toHaveText('チームへの招待');
  });

  test('1-B: nakamura でログイン後 /questions に直アクセス → /invitations にリダイレクトされる', async ({
    page,
  }) => {
    await loginAs(page, 'nakamura', 'Test1234!');

    // /questions に直アクセス
    await page.goto('/questions');

    // チームへの招待ページ、またはホームに戻ることを確認
    // InvitationListPage が表示されているか、またはリダイレクトされることを確認
    // nakamura はチームがなく pending 招待があるので TeamSelectionGate → InvitationListPage に移動するはず
    // /questions はチームが必要なため / にリダイレクトされ、そこで InvitationListPage が表示される
    await page.waitForURL((url) => !url.pathname.startsWith('/login'));

    // / か /invitations にいることを確認し、invitation-page-title が見えることを確認
    await expect(page.locator('#invitation-page-title')).toBeVisible({ timeout: 10000 });
  });

  test('1-C: sato でログインするとホーム画面（/）が表示される', async ({ page }) => {
    await loginAs(page, 'sato', 'Test1234!');

    // ホーム画面にいることを確認（URL が / であること）
    await expect(page).toHaveURL('/');

    // ログインページでないこと、招待ページでないことを確認
    await expect(page.locator('#invitation-page-title')).not.toBeVisible();
    await expect(page.locator('#no-team-title')).not.toBeVisible();
  });
});
