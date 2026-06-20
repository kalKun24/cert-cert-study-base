import type { Page } from '@playwright/test';

// テストユーザーの共通パスワード（環境変数から取得、未設定時はローカル開発用デフォルト値を使用）
const TEST_USER_PASSWORD = process.env['E2E_TEST_USER_PASSWORD'] ?? 'Test1234!';
const ADMIN_PASSWORD = process.env['E2E_ADMIN_PASSWORD'] ?? 'Admin1234!';

/** ユーザー名に応じたデフォルトパスワードを返す */
function defaultPassword(username: string): string {
  return username === 'admin' ? ADMIN_PASSWORD : TEST_USER_PASSWORD;
}

/**
 * E2E テスト用ログインヘルパー。
 * ブラウザ経由でログインフォームを操作してセッションを確立する。
 * パスワードを省略した場合は環境変数（E2E_TEST_USER_PASSWORD）のデフォルト値を使用する。
 */
export async function loginAs(page: Page, username: string, password?: string): Promise<void> {
  const pw = password ?? defaultPassword(username);
  await page.goto('/login');
  await page.getByLabel('ユーザー名').fill(username);
  await page.getByLabel('パスワード').fill(pw);
  await page.getByRole('button', { name: 'ログイン' }).click();
  // ログイン後のリダイレクトが完了するまで待機
  await page.waitForURL((url) => !url.pathname.includes('/login'));
}
