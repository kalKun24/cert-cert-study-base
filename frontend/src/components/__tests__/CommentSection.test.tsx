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
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { language: 'ja' },
  }),
}));

// react-markdown をモック（テスト環境でのESM互換性対応）
vi.mock('react-markdown', () => ({
  default: ({ children }: { children: string }) => <div>{children}</div>,
}));

// rehype-sanitize をモック
vi.mock('rehype-sanitize', () => ({
  default: {},
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
  it('マウント時に fetchComments(questionId) が呼ばれる', async () => {
    setupAuth();
    mockFetchComments.mockResolvedValue([]);

    render(<CommentSection questionId={QUESTION_ID} />);

    await waitFor(() => {
      expect(mockFetchComments).toHaveBeenCalledWith(QUESTION_ID);
    });
    // テスト終了前にローディングが完了するまで待ち、pending な非同期処理を解消する
    await waitFor(() => {
      expect(screen.queryByRole('status')).not.toBeInTheDocument();
    });
  });

  it('初回レンダー時は role="status" のローディング要素が表示される', async () => {
    setupAuth();
    mockFetchComments.mockResolvedValue([]);

    render(<CommentSection questionId={QUESTION_ID} />);

    // render の直後は isLoading=true
    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.getByRole('status').textContent).toBe('common.loading');

    // テスト終了前に非同期処理を完了させる
    await waitFor(() => {
      expect(screen.queryByRole('status')).not.toBeInTheDocument();
    });
  });

  it('fetchComments が成功するとコメントリストがレンダリングされる', async () => {
    setupAuth();
    mockFetchComments.mockResolvedValue([mockComment]);

    render(<CommentSection questionId={QUESTION_ID} />);

    await waitFor(() => {
      expect(screen.getByText(mockComment.display_name)).toBeInTheDocument();
    });
  });

  it('fetchComments が失敗すると role="alert" の要素が表示される', async () => {
    setupAuth();
    mockFetchComments.mockRejectedValue(new Error('fetch failed'));

    render(<CommentSection questionId={QUESTION_ID} />);

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
  });

  it('textarea に入力して submit すると postComment が呼ばれる', async () => {
    setupAuth();
    mockFetchComments.mockResolvedValue([]);
    mockPostComment.mockResolvedValue({
      ...mockComment,
      id: 'comment-new',
      created_by: testUser.id,
      display_name: testUser.display_name,
      body: 'テスト投稿',
    });

    render(<CommentSection questionId={QUESTION_ID} />);

    await waitFor(() => {
      expect(screen.queryByRole('status')).not.toBeInTheDocument();
    });

    const textarea = screen.getByRole('textbox', {
      name: 'comment.form.bodyLabel',
    });
    fireEvent.change(textarea, { target: { value: 'テスト投稿' } });

    const submitButton = screen.getByRole('button', {
      name: 'comment.form.submit',
    });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockPostComment).toHaveBeenCalledWith(QUESTION_ID, 'テスト投稿');
    });
  });

  it('空のコメントを送信するとバリデーションエラーが role="alert" で表示され postComment は呼ばれない', async () => {
    setupAuth();
    mockFetchComments.mockResolvedValue([]);

    render(<CommentSection questionId={QUESTION_ID} />);

    await waitFor(() => {
      expect(screen.queryByRole('status')).not.toBeInTheDocument();
    });

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
