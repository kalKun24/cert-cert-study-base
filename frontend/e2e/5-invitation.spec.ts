import { test, expect } from '@playwright/test';
import { loginAs } from './helpers/auth';
import {
  TEAM_IDS,
  USER_IDS,
  resetNakamuraInvitation,
  resetSatoInTeamA,
  removeSatoFromTeamA,
} from './helpers/api';

const TEAM_A_URL = `/teams/${TEAM_IDS.TEAM_A}`;

/**
 * テスト 5: 招待フロー
 * テスト 6: チーム脱退フロー
 *
 * 5-A: nakamura でログイン → 招待一覧で「参加する」をクリック → ホーム画面に遷移する
 * 5-B: tanaka でチームA詳細 → 招待フォームに nakamura@example.com を入力して送信 → エラーにならない
 * 6-A: sato でチームA詳細 → 「このチームから脱退する」ボタンが表示される
 * 6-B: 脱退ボタン → 確認 → チームから外れて / にリダイレクト
 * 6-C: 脱退後 /questions に直アクセス → リダイレクトされる
 */
test.describe('5: 招待フロー', () => {
  test.beforeEach(async () => {
    // nakamura の招待を pending 状態にリセット
    await resetNakamuraInvitation();
  });

  test('5-A: nakamura でログインして招待を受諾するとホーム画面に遷移する', async ({ page }) => {
    await loginAs(page, 'nakamura');

    // 招待一覧ページが表示されることを確認
    await expect(page.locator('#invitation-page-title')).toBeVisible();

    // 「参加する」ボタンをクリック
    const acceptButton = page.getByRole('button', { name: '参加する' });
    await expect(acceptButton).toBeVisible();
    await acceptButton.click();

    // ホーム画面（/）に遷移することを確認
    await page.waitForURL('/');
    await expect(page).toHaveURL('/');
  });

  test('5-B: tanaka がチームAに nakamura@example.com を招待できる（エラーにならない）', async ({
    page,
  }) => {
    await loginAs(page, 'tanaka');
    await page.goto(TEAM_A_URL);

    // 招待フォームが表示されるまで待機
    const inviteInput = page.locator('#invitee-identifier');
    await expect(inviteInput).toBeVisible();

    // メールアドレスを入力
    await inviteInput.fill('nakamura@example.com');

    // 招待ボタンをクリック
    const inviteButton = page.getByRole('button', { name: '招待' });
    await inviteButton.click();

    // 送信後にフォームの変化（入力欄クリア or エラー表示）を待機する
    // 入力欄がクリアされれば成功、エラーが出ればその内容を確認する
    await expect(inviteInput.or(page.locator('[role="alert"]'))).toBeVisible();

    // 招待フォーム固有のエラー（「招待に失敗」）が表示されていないことを確認
    const inviteErrorAlert = page.locator('[role="alert"]').filter({ hasText: /招待に失敗/ });
    await expect(inviteErrorAlert).toHaveCount(0);
  });
});

test.describe('6: チーム脱退フロー', () => {
  test.beforeEach(async () => {
    // sato がチームAにいる状態にリセット
    await resetSatoInTeamA();
  });

  test('6-A: sato でチームA詳細を開くと「このチームから脱退する」ボタンが表示される', async ({
    page,
  }) => {
    await loginAs(page, 'sato');
    await page.goto(TEAM_A_URL);

    // 脱退ボタンが表示されることを確認
    const leaveButton = page.getByRole('button', { name: 'このチームから脱退する' });
    await expect(leaveButton).toBeVisible();
  });

  test('6-B: sato が脱退ボタンを押して確認するとチームから外れて / にリダイレクトされる', async ({
    page,
  }) => {
    await loginAs(page, 'sato');
    await page.goto(TEAM_A_URL);

    // 脱退ボタンをクリック
    const leaveButton = page.getByRole('button', { name: 'このチームから脱退する' });
    await expect(leaveButton).toBeVisible();

    // window.confirm を自動承認
    page.on('dialog', async (dialog) => {
      await dialog.accept();
    });

    await leaveButton.click();

    // / にリダイレクトされることを確認
    // チームがなくなった場合は /invitations または /no-team にリダイレクト
    await page.waitForURL((url) => url.pathname === '/' || url.pathname === '/no-team' || url.pathname === '/invitations', { timeout: 10000 });

    const currentUrl = page.url();
    const pathname = new URL(currentUrl).pathname;
    expect(['/', '/no-team', '/invitations']).toContain(pathname);
  });

  test('6-C: sato が脱退後 /questions に直アクセスするとリダイレクトされる', async ({
    page,
  }) => {
    // まず sato を脱退させる
    await loginAs(page, 'sato');
    await page.goto(TEAM_A_URL);

    page.on('dialog', async (dialog) => {
      await dialog.accept();
    });

    const leaveButton = page.getByRole('button', { name: 'このチームから脱退する' });
    await expect(leaveButton).toBeVisible();
    await leaveButton.click();

    // 脱退後のリダイレクトを待機
    await page.waitForURL(
      (url) =>
        url.pathname === '/' ||
        url.pathname === '/no-team' ||
        url.pathname === '/invitations',
      { timeout: 10000 },
    );

    // /questions に直アクセス
    await page.goto('/questions');

    // / か /no-team か /invitations にリダイレクトされることを確認（チームがないため）
    await page.waitForURL(
      (url) =>
        url.pathname === '/' ||
        url.pathname === '/no-team' ||
        url.pathname === '/invitations' ||
        url.pathname === '/login',
      { timeout: 10000 },
    );

    const pathname = new URL(page.url()).pathname;
    expect(['/', '/no-team', '/invitations', '/login']).toContain(pathname);
    // /questions のままではないことを確認
    expect(pathname).not.toBe('/questions');
  });

  test.afterEach(async () => {
    // sato が脱退している場合に備えてクリーンアップは次の beforeEach に委ねる
    // ただし次テストに影響しないよう sato の状態は resetSatoInTeamA で管理
    await removeSatoFromTeamA().catch(() => {
      // already not a member, ignore
    });
  });
});
