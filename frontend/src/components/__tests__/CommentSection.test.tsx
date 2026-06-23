import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import CommentSection from '../CommentSection';
import * as commentApi from '../../utils/commentApi';
import * as AuthContextModule from '../../context/AuthContext';
import type { Comment } from '../../types/comment';
import type { AuthUser } from '../../utils/auth';

// commentApi をモック
vi.mock('../../utils/commentApi', () => ({
  fetchComments: vi.fn(),
  postComment: vi.fn(),
  updateComment: vi.fn(),
  deleteComment: vi.fn(),
}));

// AuthContext をモック
vi.mock('../../context/AuthContext', () => ({
  useAuth: vi.fn(),
}));

// react-i18next をモック（キーをそのまま返す）
// t を factory スコープの固定参照にすることで useEffect([questionId, t]) の
// 無限再実行を防ぐ（毎レンダーで新参照が生成されると setLoadError('') が
// 繰り返されエラー状態がクリアされてしまう）
vi.mock('react-i18next', () => {
  const t = (key: string) => key;
  return {
    useTranslation: () => ({
      t,
      i18n: { language: 'ja' },
    }),
  };
});

// MarkdownPreviewContent をモック（react-markdown / rehype-sanitize の ESM 互換性問題を回避）
vi.mock('../MarkdownPreviewContent', () => ({
  default: ({ value }: { value: string }) => <div>{value}</div>,
}));

const mockFetchComments = vi.mocked(commentApi.fetchComments);
const mockPostComment = vi.mocked(commentApi.postComment);
const mockUseAuth = vi.mocked(AuthContextModule.useAuth);

const testUser: AuthUser = {
  id: 'user-1',
  username: 'testuser',
  display_name: 'Test User',
  email: 'test@example.com',
  role: 'user',
  is_active: true,
};

const mockComment: Comment = {
  id: 'comment-1',
  question_id: 'question-1',
  body: 'テストコメント本文',
  created_by: 'other-user',
  display_name: 'Other User',
  created_at: '2026-01-01T10:00:00Z',
  updated_at: '2026-01-01T10:00:00Z',
};

const TEAM_ID = 'team-1';
const QUESTION_ID = 'question-1';

function setupAuth(user: AuthUser | null = testUser) {
  mockUseAuth.mockReturnValue({
    isAuthenticated: user !== null,
    user,
    token: user !== null ? 'token' : null,
    login: vi.fn(),
    logout: vi.fn(),
  });
}

afterEach(() => {
  vi.clearAllMocks();
});

describe('CommentSection', () => {
  it('マウント時に fetchComments(teamId, questionId) が呼ばれる', async () => {
    setupAuth();
    mockFetchComments.mockResolvedValue([]);

    render(<CommentSection teamId={TEAM_ID} questionId={QUESTION_ID} />);

    await waitFor(() => {
      expect(mockFetchComments).toHaveBeenCalledWith(TEAM_ID, QUESTION_ID);
    });
    await waitFor(() => {
      expect(screen.queryByRole('status')).not.toBeInTheDocument();
    });
  });

  it('初回レンダー時は role="status" のローディング要素が表示される', () => {
    setupAuth();
    // 解決しない Promise でローディング状態を維持する
    mockFetchComments.mockImplementation(() => new Promise<Comment[]>(() => {}));

    render(<CommentSection teamId={TEAM_ID} questionId={QUESTION_ID} />);

    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.getByRole('status').textContent).toBe('common.loading');
  });

  it('fetchComments が成功するとコメントリストがレンダリングされる', async () => {
    setupAuth();
    mockFetchComments.mockResolvedValue([mockComment]);

    render(<CommentSection teamId={TEAM_ID} questionId={QUESTION_ID} />);

    await waitFor(() => {
      expect(screen.getByText(mockComment.display_name)).toBeInTheDocument();
    });
  });

  it('fetchComments が失敗すると role="alert" の要素が表示される', async () => {
    setupAuth();
    // mockRejectedValue は環境によってはグローバルな unhandledRejection を引き起こすため
    // async 関数で throw する実装を使う
    mockFetchComments.mockImplementation(async () => {
      throw new Error('fetch failed');
    });

    render(<CommentSection teamId={TEAM_ID} questionId={QUESTION_ID} />);

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
  });

  it('textarea に入力して submit すると postComment が呼ばれる', async () => {
    setupAuth();
    // フォームは loading 中でも常時レンダーされるため、loading 完了を待たずに操作できる
    mockFetchComments.mockResolvedValue([]);
    mockPostComment.mockResolvedValue({
      ...mockComment,
      id: 'comment-new',
      created_by: testUser.id,
      display_name: testUser.display_name,
      body: 'テスト投稿',
    });

    render(<CommentSection teamId={TEAM_ID} questionId={QUESTION_ID} />);

    const textarea = screen.getByRole('textbox', {
      name: 'comment.form.bodyLabel',
    });
    fireEvent.change(textarea, { target: { value: 'テスト投稿' } });

    const submitButton = screen.getByRole('button', {
      name: 'comment.form.submit',
    });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockPostComment).toHaveBeenCalledWith(TEAM_ID, QUESTION_ID, 'テスト投稿');
    });
  });

  it('空のコメントを送信するとバリデーションエラーが role="alert" で表示され postComment は呼ばれない', async () => {
    setupAuth();
    // フォームは loading 中でも常時レンダーされるため、loading 完了を待たずに操作できる
    mockFetchComments.mockImplementation(() => new Promise<Comment[]>(() => {}));

    render(<CommentSection teamId={TEAM_ID} questionId={QUESTION_ID} />);

    const submitButton = screen.getByRole('button', {
      name: 'comment.form.submit',
    });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
      expect(screen.getByRole('alert').textContent).toBe(
        'comment.validation.bodyRequired',
      );
    });

    expect(mockPostComment).not.toHaveBeenCalled();
  });
});
