import { useState, useCallback, useRef } from 'react';
import CodeMirror from '@uiw/react-codemirror';
import { markdown, markdownLanguage } from '@codemirror/lang-markdown';
import { languages } from '@codemirror/language-data';
import { EditorView } from '@codemirror/view';
import { EditorSelection } from '@codemirror/state';
import { useTranslation } from 'react-i18next';
import ReactMarkdown from 'react-markdown';
import rehypeSanitize from 'rehype-sanitize';

/* =============================================================================
   型定義
   ============================================================================= */

export interface MarkdownEditorProps {
  value: string;
  onChange: (value: string) => void;
  height?: string | number;
}

type ViewMode = 'edit' | 'preview' | 'split';

/* =============================================================================
   Teal テーマ（CSS カスタムプロパティ使用）
   ============================================================================= */

const tealTheme = EditorView.theme({
  '&': {
    height: '100%',
    fontSize: '0.9375rem',
    fontFamily: "Consolas, 'Cascadia Code', 'Fira Code', 'Source Code Pro', monospace",
    backgroundColor: 'var(--color-bg-card)',
    color: 'var(--color-text-body)',
  },
  '.cm-scroller': {
    overflow: 'auto',
    height: '100%',
  },
  '.cm-content': {
    caretColor: 'var(--color-accent)',
    padding: '12px 16px',
    lineHeight: '1.7',
    minHeight: '100%',
  },
  '.cm-focused': {
    outline: 'none',
  },
  '.cm-focused .cm-cursor': {
    borderLeftColor: 'var(--color-accent)',
  },
  '.cm-selectionBackground': {
    backgroundColor: 'rgba(0, 137, 123, 0.15)',
  },
  '&.cm-focused .cm-selectionBackground': {
    backgroundColor: 'rgba(0, 137, 123, 0.2)',
  },
  '.cm-activeLine': {
    backgroundColor: 'rgba(0, 137, 123, 0.04)',
  },
  '.cm-gutters': {
    backgroundColor: 'var(--color-bg-card)',
    borderRight: '1px solid var(--color-border-default)',
    color: 'var(--color-text-muted)',
  },
  '.cm-lineNumbers .cm-gutterElement': {
    paddingLeft: '8px',
    paddingRight: '8px',
  },
  // Markdown シンタックスハイライト
  '.tok-heading': { color: '#00695c', fontWeight: '700' },
  '.tok-strong': { color: 'var(--color-text-primary)', fontWeight: '700' },
  '.tok-emphasis': { color: '#555', fontStyle: 'italic' },
  '.tok-url': { color: '#0288d1', textDecoration: 'underline' },
  '.tok-link': { color: '#0288d1' },
  '.tok-monospace': {
    color: '#b71c1c',
    backgroundColor: 'rgba(183, 28, 28, 0.06)',
    borderRadius: '3px',
    padding: '0 3px',
  },
  '.tok-meta': { color: 'var(--color-text-muted)' },
  '.tok-comment': { color: 'var(--color-text-muted)', fontStyle: 'italic' },
});

/* =============================================================================
   ツールバー挿入ヘルパー
   ============================================================================= */

interface InsertOptions {
  prefix: string;
  suffix?: string;
  defaultText?: string;
  block?: boolean;
}

function buildInsertion(
  currentValue: string,
  selectionFrom: number,
  selectionTo: number,
  options: InsertOptions,
): { newValue: string; cursorFrom: number; cursorTo: number } {
  const { prefix, suffix = '', defaultText = '', block = false } = options;
  const selected = currentValue.slice(selectionFrom, selectionTo);
  const text = selected || defaultText;

  if (block) {
    // ブロック要素: 選択範囲の前後に改行を入れる
    const before = selectionFrom > 0 ? '\n\n' : '';
    const after = '\n\n';
    const inserted = `${before}${prefix}${text}${suffix}${after}`;
    const newValue = currentValue.slice(0, selectionFrom) + inserted + currentValue.slice(selectionTo);
    const start = selectionFrom + before.length + prefix.length;
    const end = start + text.length;
    return { newValue, cursorFrom: start, cursorTo: end };
  }

  const inserted = `${prefix}${text}${suffix}`;
  const newValue = currentValue.slice(0, selectionFrom) + inserted + currentValue.slice(selectionTo);
  const start = selectionFrom + prefix.length;
  const end = start + text.length;
  return { newValue, cursorFrom: start, cursorTo: end };
}

/* =============================================================================
   MarkdownEditor コンポーネント
   ============================================================================= */

export default function MarkdownEditor({ value, onChange, height = '100%' }: MarkdownEditorProps) {
  const { t } = useTranslation();
  const [viewMode, setViewMode] = useState<ViewMode>('split');
  const editorViewRef = useRef<EditorView | null>(null);

  // CodeMirror インスタンスの参照を保持・解放
  const handleEditorCreate = useCallback((view: EditorView) => {
    editorViewRef.current = view;
  }, []);

  // プレビューモードに切り替わり CodeMirror がアンマウントされる際に参照をクリア
  const handleEditorDestroy = useCallback(() => {
    editorViewRef.current = null;
  }, []);

  // ツールバーから挿入を実行
  const insertMarkdown = useCallback(
    (options: InsertOptions) => {
      const view = editorViewRef.current;
      if (!view) {
        // エディタが非表示の場合は値を直接更新（フォールバック）
        const { newValue } = buildInsertion(value, value.length, value.length, options);
        onChange(newValue);
        return;
      }

      const { prefix, suffix = '', defaultText = '', block = false } = options;
      const state = view.state;
      const range = state.selection.main;
      const selected = state.doc.sliceString(range.from, range.to);
      const text = selected || defaultText;

      if (block) {
        // ブロック要素: 選択範囲の前後に改行を付けて range のみ置き換える
        const before = range.from > 0 ? '\n\n' : '';
        const after = '\n\n';
        const insert = `${before}${prefix}${text}${suffix}${after}`;
        view.dispatch({
          changes: { from: range.from, to: range.to, insert },
          selection: EditorSelection.range(
            range.from + before.length + prefix.length,
            range.from + before.length + prefix.length + text.length,
          ),
          userEvent: 'input',
        });
      } else {
        // インライン要素: 選択範囲のみ置き換える
        const insert = `${prefix}${text}${suffix}`;
        view.dispatch({
          changes: { from: range.from, to: range.to, insert },
          selection: EditorSelection.range(
            range.from + prefix.length,
            range.from + prefix.length + text.length,
          ),
          userEvent: 'input',
        });
      }
      view.focus();
    },
    [value, onChange],
  );

  const handleBold = () => insertMarkdown({ prefix: '**', suffix: '**', defaultText: '太字テキスト' });
  const handleItalic = () => insertMarkdown({ prefix: '*', suffix: '*', defaultText: '斜体テキスト' });
  const handleLink = () => insertMarkdown({ prefix: '[', suffix: '](URL)', defaultText: 'リンクテキスト' });
  const handleTable = () =>
    insertMarkdown({
      prefix: '| 列1 | 列2 | 列3 |\n| --- | --- | --- |\n| セル | セル | セル |',
      block: true,
      defaultText: '',
    });
  const handleCodeBlock = () =>
    insertMarkdown({ prefix: '```\n', suffix: '\n```', defaultText: 'コード', block: true });
  const handleChecklist = () =>
    insertMarkdown({ prefix: '- [ ] ', defaultText: 'タスク', block: true });

  const heightValue = typeof height === 'number' ? `${height}px` : height;

  return (
    <div
      className="md-editor"
      style={{ height: heightValue }}
      data-testid="markdown-editor"
    >
      {/* ツールバー */}
      <div className="md-editor-toolbar" role="toolbar" aria-label="Markdownツールバー">
        <div className="md-editor-toolbar-group">
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleBold}
            title={t('editor.toolbar.bold')}
            aria-label={t('editor.toolbar.bold')}
          >
            <strong>B</strong>
          </button>
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleItalic}
            title={t('editor.toolbar.italic')}
            aria-label={t('editor.toolbar.italic')}
          >
            <em>I</em>
          </button>
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleLink}
            title={t('editor.toolbar.link')}
            aria-label={t('editor.toolbar.link')}
          >
            🔗
          </button>
        </div>

        <div className="md-editor-toolbar-separator" aria-hidden="true" />

        <div className="md-editor-toolbar-group">
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleTable}
            title={t('editor.toolbar.table')}
            aria-label={t('editor.toolbar.table')}
          >
            ⊞
          </button>
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleCodeBlock}
            title={t('editor.toolbar.codeBlock')}
            aria-label={t('editor.toolbar.codeBlock')}
          >
            {'</>'}
          </button>
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleChecklist}
            title={t('editor.toolbar.checklist')}
            aria-label={t('editor.toolbar.checklist')}
          >
            ☑
          </button>
        </div>

        {/* ビューモード切り替え */}
        <div className="md-editor-toolbar-mode" role="group" aria-label="表示モード">
          {(['edit', 'split', 'preview'] as const).map((mode) => (
            <button
              key={mode}
              type="button"
              className={`md-editor-mode-btn${viewMode === mode ? ' md-editor-mode-btn--active' : ''}`}
              onClick={() => setViewMode(mode)}
              aria-pressed={viewMode === mode}
              aria-label={t(`editor.viewMode.${mode}`)}
            >
              {t(`editor.viewMode.${mode}`)}
            </button>
          ))}
        </div>
      </div>

      {/* エディタ本体 */}
      <div
        className={`md-editor-body md-editor-body--${viewMode}`}
        data-testid="markdown-editor-body"
      >
        {/* 編集エリア */}
        {viewMode !== 'preview' && (
          <div className="md-editor-pane md-editor-pane--edit">
            <CodeMirror
              value={value}
              onChange={onChange}
              onCreateEditor={handleEditorCreate}
              onDestroy={handleEditorDestroy}
              extensions={[
                markdown({ base: markdownLanguage, codeLanguages: languages }),
                tealTheme,
                EditorView.lineWrapping,
              ]}
              basicSetup={{
                lineNumbers: false,
                foldGutter: false,
                dropCursor: true,
                allowMultipleSelections: true,
                indentOnInput: true,
                bracketMatching: true,
                closeBrackets: true,
                autocompletion: false,
                rectangularSelection: false,
                crosshairCursor: false,
                highlightActiveLine: true,
                highlightSelectionMatches: false,
                closeBracketsKeymap: true,
                defaultKeymap: true,
                searchKeymap: false,
                historyKeymap: true,
                foldKeymap: false,
                completionKeymap: false,
                lintKeymap: false,
              }}
              style={{ height: '100%' }}
              aria-label="Markdownエディタ"
            />
          </div>
        )}

        {/* プレビューエリア */}
        {viewMode !== 'edit' && (
          <div
            className={`md-editor-pane md-editor-pane--preview${viewMode === 'split' ? ' md-editor-pane--preview-split' : ''}`}
            aria-label="プレビュー"
          >
            <div className="md-editor-preview-content markdown-body">
              {value.trim() ? (
                <ReactMarkdown rehypePlugins={[rehypeSanitize]}>{value}</ReactMarkdown>
              ) : (
                <p className="md-editor-preview-empty">{t('editor.preview.empty')}</p>
              )}
            </div>
          </div>
        )}
      </div>

      {/* モバイル: プレビュータブ（スプリットモード時は下にプレビュー） */}
    </div>
  );
}
