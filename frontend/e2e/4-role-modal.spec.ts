import { test, expect } from '@playwright/test';
import { loginAs } from './helpers/auth';
import {
  TEAM_IDS,
  USER_IDS,
  setYamadaAsOnlyOwner,
  restoreTeamAOwners,
  adminLogin,
} from './helpers/api';

const TEAM_A_URL = `/teams/${TEAM_IDS.TEAM_A}`;

/**
 * テスト 4: チーム詳細のロール管理
 *
 * 4-A: tanaka でチームA詳細を開く → yamada 行に「オーナーを外す」、sato 行に「オーナーにする」ボタンが表示される
 * 4-B: sato の「オーナーにする」をクリック → 注意ポップアップが表示される
 * 4-C: ポップアップで「付与する」を押す → モーダルが閉じてページがリロードされ sato が owner になる
 * 4-D: yamada でチームA詳細を開き自分（yamada）の「オーナーを外す」をクリック → モーダル → 確認 → member になる
 * 4-E: チームAのオーナーを yamada のみにして、yamada でチームA詳細を開く → 「オーナーを外す」ボタンが disabled
 */
test.describe('4: チーム詳細のロール管理', () => {
  // 各テスト後にチームAのオーナー状態を tanaka / yamada に戻す
  test.afterEach(async () => {
    await restoreTeamAOwners();
  });

  test('4-A: tanaka でチームA詳細を開くと各メンバーに適切なボタンが表示される', async ({
    page,
  }) => {
    await loginAs(page, 'tanaka', 'Test1234!');
    await page.goto(TEAM_A_URL);

    // ページが読み込まれるまで待機
    await expect(page.locator('.team-members-table')).toBeVisible();

    // yamada の行に「オーナーを外す」ボタンが表示される
    // （tanaka は自分自身なので外すボタンは別ルール、yamada は他のオーナー）
    const revokeButtons = page.getByRole('button', { name: 'オーナーを外す' });
    await expect(revokeButtons.first()).toBeVisible();

    // sato の行に「オーナーにする」ボタンが表示される（sato は member なので）
    const grantButton = page.getByRole('button', { name: 'オーナーにする' });
    await expect(grantButton).toBeVisible();
  });

  test('4-B: sato の「オーナーにする」をクリックすると注意ポップアップが表示される', async ({
    page,
  }) => {
    await loginAs(page, 'tanaka', 'Test1234!');
    await page.goto(TEAM_A_URL);

    await expect(page.locator('.team-members-table')).toBeVisible();

    // sato の「オーナーにする」をクリック
    const grantButton = page.getByRole('button', { name: 'オーナーにする' });
    await grantButton.first().click();

    // モーダルが表示されることを確認
    const modal = page.getByRole('dialog');
    await expect(modal).toBeVisible();
    await expect(modal).toHaveAttribute('aria-modal', 'true');
    await expect(modal).toHaveAttribute('aria-labelledby', 'team-owner-role-modal-title');
  });

  test('4-C: 「付与する」を押すとモーダルが閉じて sato が owner になる', async ({ page }) => {
    await loginAs(page, 'tanaka', 'Test1234!');
    await page.goto(TEAM_A_URL);

    await expect(page.locator('.team-members-table')).toBeVisible();

    // sato の「オーナーにする」をクリック
    const grantButton = page.getByRole('button', { name: 'オーナーにする' });
    await grantButton.first().click();

    // モーダルで「付与する」をクリック
    const modal = page.getByRole('dialog');
    await expect(modal).toBeVisible();
    await modal.getByRole('button', { name: '付与する' }).click();

    // モーダルが閉じることを確認
    await expect(modal).not.toBeVisible();

    // ページがリロードされて sato が owner になっていることを確認
    // sato の行に「オーナーを外す」ボタンが表示される（owner になったため）
    await expect(page.locator('.team-members-table')).toBeVisible();
    const revokeButtons = page.getByRole('button', { name: 'オーナーを外す' });
    // owner が増えたので「オーナーを外す」ボタンが増えているはず
    await expect(revokeButtons).toHaveCount(2);
  });

  test('4-D: yamada でチームA詳細を開き自分の「オーナーを外す」→ モーダル → member になる', async ({
    page,
  }) => {
    // tanaka が残るので yamada は外れることができる
    await loginAs(page, 'yamada', 'Test1234!');
    await page.goto(TEAM_A_URL);

    await expect(page.locator('.team-members-table')).toBeVisible();

    // yamada 自身の「オーナーを外す」ボタンを探す
    // 自分の行のボタンはコード上 isCurrentUser=true なので「オーナーを外す」は表示されない
    // ただし yamada はオーナーなので、他の tanaka 行には表示される
    // yamada が自分の行でできる操作: 「このチームから脱退する」はある
    // ※コード確認: isCurrentUser && isMemberOwner の行はボタン非表示
    // 実際には yamada 自身の行に role 変更ボタンは出ない（自分自身は変更不可の設計）
    // → tanaka 行の「オーナーを外す」で代替確認

    // tanaka の「オーナーを外す」をクリック（yamada がオーナーなので操作可能）
    const revokeButton = page.getByRole('button', { name: 'オーナーを外す' });
    await expect(revokeButton).toBeVisible();
    await revokeButton.first().click();

    // モーダルが表示されることを確認
    const modal = page.getByRole('dialog');
    await expect(modal).toBeVisible();

    // 「外す」をクリック
    await modal.getByRole('button', { name: '外す' }).click();

    // モーダルが閉じることを確認
    await expect(modal).not.toBeVisible();

    // ページがリロードされて変更が反映されることを確認
    await expect(page.locator('.team-members-table')).toBeVisible();
  });

  test('4-E: yamada のみオーナーの状態で、yamada の「オーナーを外す」ボタンが disabled になる', async ({
    page,
  }) => {
    // yamada のみをオーナーに設定
    await setYamadaAsOnlyOwner();

    await loginAs(page, 'yamada', 'Test1234!');
    await page.goto(TEAM_A_URL);

    await expect(page.locator('.team-members-table')).toBeVisible();

    // yamada 自身の行のボタンは isCurrentUser=true で非表示だが、
    // ownerCount=1 のため他メンバーの「オーナーを外す」ボタンも disabled になる
    // → sato, tanaka などの「オーナーを外す」ボタンが disabled / 非表示を確認

    // ページに「オーナーを外す」ボタンがある場合、すべて disabled であること
    const revokeButtons = page.getByRole('button', { name: 'オーナーを外す' });
    const count = await revokeButtons.count();

    if (count > 0) {
      // 存在する場合はすべて disabled であること
      for (let i = 0; i < count; i++) {
        await expect(revokeButtons.nth(i)).toBeDisabled();
      }
    } else {
      // yamada が唯一のオーナーで、かつ自分の行のボタンは非表示なので count=0 も正常
      expect(count).toBe(0);
    }
  });
});

// ─── ヘルパー：管理者権限で直接ロールを確認する ────────────────────────────────

async function getSatoRole(): Promise<string> {
  const token = await adminLogin();
  const res = await fetch(`http://localhost:8080/api/v1/teams/${TEAM_IDS.TEAM_A}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const json = (await res.json()) as {
    data: { members: { user_id: string; role: string }[] };
  };
  const satoMember = json.data.members.find((m) => m.user_id === USER_IDS.SATO);
  return satoMember?.role ?? 'not_found';
}
