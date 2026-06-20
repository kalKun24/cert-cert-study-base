/**
 * E2E テスト用 API ヘルパー
 * テストデータのセットアップ・クリーンアップに使用する
 *
 * 認証情報は環境変数から取得します。
 * ローカル開発時は frontend/.env.e2e.local を作成してください（.gitignore 対象）。
 * （frontend/.env.e2e.example を参照）
 */

export const API_BASE = 'http://localhost:8080/api/v1';

// E2E テスト用認証情報（環境変数から取得）
const ADMIN_USERNAME = process.env['E2E_ADMIN_USERNAME'] ?? 'admin';
const ADMIN_PASSWORD = process.env['E2E_ADMIN_PASSWORD'] ?? 'Admin1234!';
const TEST_USER_PASSWORD = process.env['E2E_TEST_USER_PASSWORD'] ?? 'Test1234!';

// ユーザー / チームの固定 ID
export const USER_IDS = {
  TANAKA: '822cc8b3-ccd3-47bf-9227-7757dd5e2156',
  SUZUKI: '9df17ea3-ea99-4c0c-8efa-ca72a6ec6c28',
  YAMADA: 'ea061b65-4a15-4a95-93c6-9032f16a9ad5',
  SATO: 'e3c3a66f-128e-412c-83d4-8cfb45ac78ec',
  NAKAMURA: 'f5ba11ec-693f-43a3-afde-f3098256b5b2',
} as const;

export const TEAM_IDS = {
  TEAM_A: 'bccf1c3a-e480-4c9b-9c7a-52968c7911ce',
  TEAM_B: 'b7be94df-1778-4e92-9942-d21931356ffd',
} as const;

// ─────────────────────────────────────────────────────────────────────────────
// 認証
// ─────────────────────────────────────────────────────────────────────────────

interface LoginResponse {
  data: {
    token: string;
  };
  error: string | null;
}

/** 管理者としてログインしてトークンを取得する */
export async function adminLogin(): Promise<string> {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: ADMIN_USERNAME, password: ADMIN_PASSWORD }),
  });
  if (!res.ok) {
    throw new Error(`adminLogin failed: HTTP ${res.status}`);
  }
  const json = (await res.json()) as LoginResponse;
  return json.data.token;
}

/** 任意のユーザーとしてログインしてトークンを取得する */
export async function loginAs(username: string, password?: string): Promise<string> {
  const pw = password ?? TEST_USER_PASSWORD;
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password: pw }),
  });
  if (!res.ok) {
    throw new Error(`loginAs(${username}) failed: HTTP ${res.status}`);
  }
  const json = (await res.json()) as LoginResponse;
  return json.data.token;
}

// ─────────────────────────────────────────────────────────────────────────────
// 招待管理
// ─────────────────────────────────────────────────────────────────────────────

interface Invitation {
  id: string;
  team_id: string;
  invitee_user_id: string;
  status: 'pending' | 'accepted' | 'rejected';
}

interface InvitationsResponse {
  data: Invitation[];
  error: string | null;
}

interface InvitationResponse {
  data: Invitation;
  error: string | null;
}

/** 自分宛ての招待一覧を取得する */
async function getMyInvitations(token: string): Promise<Invitation[]> {
  const res = await fetch(`${API_BASE}/me/invitations`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) {
    throw new Error(`getMyInvitations failed: ${res.status}`);
  }
  const json = (await res.json()) as InvitationsResponse;
  return json.data ?? [];
}

/** チームに招待を送信する（オーナートークンが必要） */
async function inviteToTeam(
  token: string,
  teamId: string,
  identifier: string,
): Promise<Invitation> {
  const res = await fetch(`${API_BASE}/teams/${teamId}/invitations`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ identifier }),
  });
  if (!res.ok) {
    throw new Error(`inviteToTeam failed: ${res.status} ${await res.text()}`);
  }
  const json = (await res.json()) as InvitationResponse;
  return json.data;
}

/** 招待に応答する */
async function respondInvitation(
  token: string,
  invitationId: string,
  status: 'accepted' | 'rejected',
): Promise<void> {
  const res = await fetch(`${API_BASE}/invitations/${invitationId}`, {
    method: 'PUT',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ status }),
  });
  if (!res.ok) {
    throw new Error(`respondInvitation failed: ${res.status} ${await res.text()}`);
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// メンバー管理
// ─────────────────────────────────────────────────────────────────────────────

interface TeamMember {
  user_id: string;
  role: 'owner' | 'member';
}

interface TeamDetail {
  members: TeamMember[];
}

interface TeamDetailResponse {
  data: TeamDetail;
  error: string | null;
}

/** チームの詳細（メンバー一覧）を取得する */
async function getTeamDetail(token: string, teamId: string): Promise<TeamDetail> {
  const res = await fetch(`${API_BASE}/teams/${teamId}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) {
    throw new Error(`getTeamDetail failed: ${res.status}`);
  }
  const json = (await res.json()) as TeamDetailResponse;
  return json.data;
}

/** メンバーのロールを変更する */
async function changeMemberRole(
  token: string,
  teamId: string,
  userId: string,
  role: 'owner' | 'member',
): Promise<void> {
  const res = await fetch(`${API_BASE}/teams/${teamId}/members/${userId}`, {
    method: 'PUT',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ role }),
  });
  if (!res.ok) {
    throw new Error(`changeMemberRole failed: ${res.status} ${await res.text()}`);
  }
}

/** チームからメンバーを除外する */
async function removeMember(token: string, teamId: string, userId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/teams/${teamId}/members/${userId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok && res.status !== 404) {
    throw new Error(`removeMember failed: ${res.status}`);
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// 公開ヘルパー関数
// ─────────────────────────────────────────────────────────────────────────────

/**
 * チームA に sato を再追加する（招待→受諾）。
 * テスト 6-A / 6-B の事前セットアップに使用する。
 */
export async function resetSatoInTeamA(): Promise<void> {
  const adminToken = await adminLogin();
  const tanakaToken = await loginAs('tanaka');
  const satoToken = await loginAs('sato');

  // すでに sato がチームAのメンバーなら何もしない
  const team = await getTeamDetail(adminToken, TEAM_IDS.TEAM_A);
  const isAlreadyMember = team.members.some((m) => m.user_id === USER_IDS.SATO);
  if (isAlreadyMember) return;

  // tanaka でチームAに sato を招待
  await inviteToTeam(tanakaToken, TEAM_IDS.TEAM_A, USER_IDS.SATO);

  // sato の pending 招待を受諾
  const satoInvitations = await getMyInvitations(satoToken);
  const pending = satoInvitations.find(
    (inv) => inv.team_id === TEAM_IDS.TEAM_A && inv.status === 'pending',
  );
  if (pending) {
    await respondInvitation(satoToken, pending.id, 'accepted');
  }
}

/**
 * nakamura へのチームB招待を pending 状態にリセットする。
 * テスト 5-A の事前セットアップに使用する。
 */
export async function resetNakamuraInvitation(): Promise<void> {
  const adminToken = await adminLogin();
  const suzukiToken = await loginAs('suzuki');
  const nakamuraToken = await loginAs('nakamura');

  // nakamura がすでにチームBのメンバーであれば除外する
  const team = await getTeamDetail(adminToken, TEAM_IDS.TEAM_B);
  const isAlreadyMember = team.members.some((m) => m.user_id === USER_IDS.NAKAMURA);
  if (isAlreadyMember) {
    await removeMember(adminToken, TEAM_IDS.TEAM_B, USER_IDS.NAKAMURA);
  }

  // 既存の pending 招待を確認（あれば流用、なければ再送）
  const nakamuraInvitations = await getMyInvitations(nakamuraToken);
  const hasPending = nakamuraInvitations.some(
    (inv) => inv.team_id === TEAM_IDS.TEAM_B && inv.status === 'pending',
  );
  if (!hasPending) {
    await inviteToTeam(suzukiToken, TEAM_IDS.TEAM_B, USER_IDS.NAKAMURA);
  }
}

/**
 * チームAのオーナーを yamada のみに変更する。
 * テスト 4-E の事前セットアップに使用する。
 * tanaka を member に降格し yamada だけ owner にする。
 */
export async function setYamadaAsOnlyOwner(): Promise<void> {
  const adminToken = await adminLogin();
  const team = await getTeamDetail(adminToken, TEAM_IDS.TEAM_A);

  // yamada を owner に（まだ owner でない場合）
  const yamadaMember = team.members.find((m) => m.user_id === USER_IDS.YAMADA);
  if (yamadaMember && yamadaMember.role !== 'owner') {
    await changeMemberRole(adminToken, TEAM_IDS.TEAM_A, USER_IDS.YAMADA, 'owner');
  }

  // yamada 以外の owner を member に降格
  for (const m of team.members) {
    if (m.user_id !== USER_IDS.YAMADA && m.role === 'owner') {
      await changeMemberRole(adminToken, TEAM_IDS.TEAM_A, m.user_id, 'member');
    }
  }
}

/**
 * チームAのオーナーを tanaka と yamada の両方に戻す。
 * テスト 4-E の事後クリーンアップに使用する。
 */
export async function restoreTeamAOwners(): Promise<void> {
  const adminToken = await adminLogin();

  // tanaka と yamada を owner に設定
  await changeMemberRole(adminToken, TEAM_IDS.TEAM_A, USER_IDS.TANAKA, 'owner');
  await changeMemberRole(adminToken, TEAM_IDS.TEAM_A, USER_IDS.YAMADA, 'owner');
  // sato は member のままにする
  await changeMemberRole(adminToken, TEAM_IDS.TEAM_A, USER_IDS.SATO, 'member');
}

/**
 * sato の is_team_owner を指定の状態にリセットする。
 * テスト 3-A〜3-D の後処理に使用する。
 */
export async function resetSatoTeamOwnerStatus(isTeamOwner: boolean): Promise<void> {
  const adminToken = await adminLogin();
  const res = await fetch(`${API_BASE}/users/${USER_IDS.SATO}/team-owner`, {
    method: 'PUT',
    headers: {
      Authorization: `Bearer ${adminToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      is_team_owner: isTeamOwner,
      max_teams: isTeamOwner ? 1 : undefined,
    }),
  });
  if (!res.ok) {
    throw new Error(`resetSatoTeamOwnerStatus failed: ${res.status}`);
  }
}

/**
 * sato をチームA から除外する。
 * テスト 6-B の後処理に使用する（脱退テストが成功した場合は不要だが冪等に）。
 */
export async function removeSatoFromTeamA(): Promise<void> {
  const adminToken = await adminLogin();
  await removeMember(adminToken, TEAM_IDS.TEAM_A, USER_IDS.SATO);
}

// ─────────────────────────────────────────────────────────────────────────────
// タグ管理
// ─────────────────────────────────────────────────────────────────────────────

interface Tag {
  id: string;
  team_id: string;
  name: string;
}

interface TagResponse {
  data: Tag;
  error: string | null;
}

interface TagListResponse {
  data: Tag[];
  error: string | null;
}

/** チームのタグ一覧を取得する */
export async function listTagsViaApi(token: string, teamId: string): Promise<Tag[]> {
  const res = await fetch(`${API_BASE}/teams/${teamId}/tags`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) {
    throw new Error(`listTagsViaApi failed: ${res.status}`);
  }
  const json = (await res.json()) as TagListResponse;
  return json.data ?? [];
}

/** チームにタグを作成してタグIDを返す */
export async function createTagViaApi(token: string, teamId: string, name: string): Promise<string> {
  const res = await fetch(`${API_BASE}/teams/${teamId}/tags`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ name }),
  });
  if (!res.ok) {
    throw new Error(`createTagViaApi failed: ${res.status} ${await res.text()}`);
  }
  const json = (await res.json()) as TagResponse;
  return json.data.id;
}

/** チームのタグを削除する */
export async function deleteTagViaApi(token: string, teamId: string, tagId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/teams/${teamId}/tags/${tagId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok && res.status !== 404) {
    throw new Error(`deleteTagViaApi failed: ${res.status}`);
  }
}

/**
 * チームのタグを全削除する（テスト後クリーンアップ用）。
 * 問題に使用中のタグは削除できないため、409 は無視する。
 */
export async function cleanupTeamTags(teamId: string): Promise<void> {
  const adminToken = await adminLogin();
  const tags = await listTagsViaApi(adminToken, teamId);
  await Promise.allSettled(
    tags.map((tag) => deleteTagViaApi(adminToken, teamId, tag.id)),
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// 問題管理
// ─────────────────────────────────────────────────────────────────────────────

interface Question {
  id: string;
  team_id: string;
  title: string;
  status: string;
}

interface QuestionResponse {
  data: Question;
  error: string | null;
}

interface QuestionListResponse {
  data: {
    items: Question[];
    total: number;
  };
  error: string | null;
}

/** チームに問題を作成してIDを返す */
export async function createQuestionViaApi(
  token: string,
  teamId: string,
  title: string,
): Promise<string> {
  const res = await fetch(`${API_BASE}/teams/${teamId}/questions`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ title, body: '## 問題\nE2Eテスト用問題文' }),
  });
  if (!res.ok) {
    throw new Error(`createQuestionViaApi failed: ${res.status} ${await res.text()}`);
  }
  const json = (await res.json()) as QuestionResponse;
  return json.data.id;
}

/** チームの問題一覧を取得する（認可チェック用） */
export async function listQuestionsViaApi(
  token: string,
  teamId: string,
): Promise<Response> {
  return fetch(`${API_BASE}/teams/${teamId}/questions`, {
    headers: { Authorization: `Bearer ${token}` },
  });
}

/** チームの問題を削除する */
export async function deleteQuestionViaApi(
  token: string,
  teamId: string,
  questionId: string,
): Promise<void> {
  const res = await fetch(`${API_BASE}/teams/${teamId}/questions/${questionId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok && res.status !== 404) {
    throw new Error(`deleteQuestionViaApi failed: ${res.status}`);
  }
}

/** チームの問題を全削除する（テスト後クリーンアップ用） */
export async function cleanupTeamQuestions(teamId: string): Promise<void> {
  const adminToken = await adminLogin();
  // admin はチームメンバーである必要があるため、tanaka のトークンで削除
  const tanakaToken = await loginAs('tanaka');
  const res = await listQuestionsViaApi(
    teamId === TEAM_IDS.TEAM_A ? tanakaToken : adminToken,
    teamId,
  );
  if (!res.ok) return;
  const json = (await res.json()) as QuestionListResponse;
  const questions = json.data?.items ?? [];
  const token = teamId === TEAM_IDS.TEAM_A ? tanakaToken : await loginAs('suzuki');
  await Promise.allSettled(
    questions.map((q) => deleteQuestionViaApi(token, teamId, q.id)),
  );
}
