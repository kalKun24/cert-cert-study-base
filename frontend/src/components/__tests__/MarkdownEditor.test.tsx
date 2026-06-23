import { render, screen, fireEvent } from '@testing-library/react';
import MarkdownEditor from '../MarkdownEditor';

// @uiw/react-codemirror は jsdom 環境で動作しないためモック
vi.mock('@uiw/react-codemirror', () => ({
  default: ({
    value,
    onChange,
    'aria-label': ariaLabel,
  }: {
    value: string;
    onChange: (value: string) => void;
    'aria-label'?: string;
  }) => (
    <textarea
      aria-label={ariaLabel ?? 'Markdownエディタ'}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      data-testid="codemirror-mock"
    />
  ),
}));

// CodeMirror 拡張モジュールのモック
vi.mock('@codemirror/lang-markdown', () => ({
  markdown: vi.fn(() => []),
  markdownLanguage: {},
}));

vi.mock('@codemirror/language-data', () => ({
  languages: [],
}));

vi.mock('@codemirror/view', () => ({
  EditorView: {
    theme: vi.fn(() => []),
    lineWrapping: [],
  },
}));

vi.mock('@codemirror/state', () => ({
  EditorSelection: {
    range: vi.fn(),
  },
}));

// react-i18next をモック（キーをそのまま返す）
vi.mock('react-i18next', () => {
  const t = (key: string) => key;
  return {
    useTranslation: () => ({
      t,
      i18n: { language: 'ja' },
    }),
  };
});

// react-markdown をモック（ESM 互換性）
vi.mock('react-markdown', () => ({
  default: ({ children }: { children: string }) => <div data-testid="markdown-preview">{children}</div>,
}));

// rehype-sanitize をモック
vi.mock('rehype-sanitize', () => ({
  default: {},
}));

const NOOP = () => {};

describe('MarkdownEditor', () => {
  it('コンポーネントがレンダリングされること', () => {
    render(<MarkdownEditor value="" onChange={NOOP} />);

    expect(screen.getByTestId('markdown-editor')).toBeInTheDocument();
  });

  it('ツールバーの各ボタンが存在すること', () => {
    render(<MarkdownEditor value="" onChange={NOOP} />);

    expect(screen.getByLabelText('editor.toolbar.bold')).toBeInTheDocument();
    expect(screen.getByLabelText('editor.toolbar.italic')).toBeInTheDocument();
    expect(screen.getByLabelText('editor.toolbar.link')).toBeInTheDocument();
    expect(screen.getByLabelText('editor.toolbar.table')).toBeInTheDocument();
    expect(screen.getByLabelText('editor.toolbar.codeBlock')).toBeInTheDocument();
    expect(screen.getByLabelText('editor.toolbar.checklist')).toBeInTheDocument();
  });

  it('ビューモードの切り替えボタンが 3 つ存在すること', () => {
    render(<MarkdownEditor value="" onChange={NOOP} />);

    expect(screen.getByLabelText('editor.viewMode.edit')).toBeInTheDocument();
    expect(screen.getByLabelText('editor.viewMode.split')).toBeInTheDocument();
    expect(screen.getByLabelText('editor.viewMode.preview')).toBeInTheDocument();
  });

  it('デフォルトのビューモードは split であること', () => {
    render(<MarkdownEditor value="" onChange={NOOP} />);

    const splitBtn = screen.getByLabelText('editor.viewMode.split');
    expect(splitBtn).toHaveAttribute('aria-pressed', 'true');
  });

  it('value が CodeMirror モックに渡されること', () => {
    const testValue = '# テスト見出し\n\nサンプルテキスト';
    render(<MarkdownEditor value={testValue} onChange={NOOP} />);

    const editor = screen.getByTestId('codemirror-mock') as HTMLTextAreaElement;
    expect(editor.value).toBe(testValue);
  });

  it('value がプレビューエリアに表示されること（split モード）', () => {
    const testValue = '# プレビューテスト';
    render(<MarkdownEditor value={testValue} onChange={NOOP} />);

    const preview = screen.getByTestId('markdown-preview');
    expect(preview).toBeInTheDocument();
    expect(preview.textContent).toBe(testValue);
  });

  it('プレビューモードに切り替えるとエディタが非表示になること', () => {
    render(<MarkdownEditor value="テスト" onChange={NOOP} />);

    const previewModeBtn = screen.getByLabelText('editor.viewMode.preview');
    fireEvent.click(previewModeBtn);

    expect(screen.queryByTestId('codemirror-mock')).not.toBeInTheDocument();
    expect(screen.getByTestId('markdown-preview')).toBeInTheDocument();
    expect(previewModeBtn).toHaveAttribute('aria-pressed', 'true');
  });

  it('編集モードに切り替えるとプレビューが非表示になること', () => {
    render(<MarkdownEditor value="テスト" onChange={NOOP} />);

    const editModeBtn = screen.getByLabelText('editor.viewMode.edit');
    fireEvent.click(editModeBtn);

    expect(screen.getByTestId('codemirror-mock')).toBeInTheDocument();
    expect(screen.queryByTestId('markdown-preview')).not.toBeInTheDocument();
    expect(editModeBtn).toHaveAttribute('aria-pressed', 'true');
  });

  it('空の value でプレビューに「プレビューするテキストがありません」メッセージが表示されること', () => {
    render(<MarkdownEditor value="" onChange={NOOP} />);

    // split モードではプレビューが表示される
    expect(screen.getByText('editor.preview.empty')).toBeInTheDocument();
  });

  it('CodeMirror の onChange が呼び出されると onChange が発火すること', () => {
    const handleChange = vi.fn();
    render(<MarkdownEditor value="" onChange={handleChange} />);

    const editor = screen.getByTestId('codemirror-mock');
    fireEvent.change(editor, { target: { value: '新しいテキスト' } });

    expect(handleChange).toHaveBeenCalledWith('新しいテキスト');
  });
});
