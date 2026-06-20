import { test, expect } from '@playwright/test';
import { loginAs } from './helpers/auth';

/**
 * テスト 2: Topbar のチーム切り替え
 *
 * 2-A: yamada でログイン → Topbar のプルダウンにチームA・チームBの両方が出る
 * 2-B: yamada でプルダウンを切り替え → 別チームが選択された状態になる
 * 2-C: tanaka でログイン → Topbar に「チームを作成」リンクが表示される（is_team_owner=true）
 * 2-D: sato でログイン → 「チームを作成」リンクが表示されない（is_team_owner=false）
 */
test.describe('2: Topbar のチーム切り替え', () => {
  test('2-A: yamada でログインするとプルダウンにチームAとチームBが表示される', async ({
    page,
  }) => {
    await loginAs(page, 'yamada');

    const teamSelect = page.locator('#topbar-team-select');
    await expect(teamSelect).toBeVisible();

    // チームAとチームBの両方がオプションに存在することを確認
    const options = teamSelect.locator('option');
    const optionTexts = await options.allTextContents();

    const hasTeamA = optionTexts.some((text) => text.includes('CISSP勉強チームA'));
    const hasTeamB = optionTexts.some((text) => text.includes('情報安全確保支援士チームB'));

    expect(hasTeamA).toBe(true);
    expect(hasTeamB).toBe(true);
  });

  test('2-B: yamada でプルダウンを切り替えると別チームが選択される', async ({ page }) => {
    await loginAs(page, 'yamada');

    const teamSelect = page.locator('#topbar-team-select');
    await expect(teamSelect).toBeVisible();

    // 現在選択中の値を取得
    const currentValue = await teamSelect.inputValue();

    // オプション一覧を取得してもう一方のチームを選択
    const options = teamSelect.locator('option');
    const allValues = await options.evaluateAll((opts: HTMLOptionElement[]) =>
      opts.map((o) => o.value),
    );

    const otherValue = allValues.find((v) => v !== currentValue);
    expect(otherValue).toBeDefined();

    // 別チームを選択
    await teamSelect.selectOption(otherValue as string);

    // 選択値が変わったことを確認
    await expect(teamSelect).toHaveValue(otherValue as string);
  });

  test('2-C: tanaka でログインすると「チームを作成」リンクが表示される', async ({ page }) => {
    await loginAs(page, 'tanaka');

    // is_team_owner=true なのでリンクが表示されるはず
    const createTeamLink = page.getByRole('link', { name: 'チームを作成' });
    await expect(createTeamLink).toBeVisible();
  });

  test('2-D: sato でログインすると「チームを作成」リンクが表示されない', async ({ page }) => {
    await loginAs(page, 'sato');

    // is_team_owner=false なのでリンクは表示されないはず
    const createTeamLink = page.getByRole('link', { name: 'チームを作成' });
    await expect(createTeamLink).not.toBeVisible();
  });
});
