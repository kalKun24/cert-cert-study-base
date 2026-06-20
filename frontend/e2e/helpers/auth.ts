import type { Page } from '@playwright/test';

/**
 * E2E テスト用ログインヘルパー。
 * ブラウザ経由でログインフォームを操作してセッションを確立する。
 */
export async function loginAs(page: Page, username: string, password: string): Promise<void> {
  await page.goto('/login');
  await page.getByLabel('ユーザー名').fill(username);
  await page.getByLabel('パスワード').fill(password);
  await page.getByRole('button', { name: 'ログイン' }).click();
  // ログイン後のリダイレクトが完了するまで待機
  await page.waitForURL((url) => !url.pathname.includes('/login'));
}
