import { useState, useCallback, useRef, useEffect, useMemo } from 'react';
import CodeMirror from '@uiw/react-codemirror';
import { markdown, markdownLanguage } from '@codemirror/lang-markdown';
import { languages } from '@codemirror/language-data';
import { EditorView, keymap } from '@codemirror/view';
import { EditorSelection, Prec } from '@codemirror/state';
import { useTranslation } from 'react-i18next';
import MarkdownPreviewContent from './MarkdownPreviewContent';
import 'highlight.js/styles/github-dark-dimmed.css';
import 'katex/dist/katex.min.css';

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

/** 挿入文字列と選択範囲オフセットを計算する純粋ヘルパー */
function buildInsertString(
  text: string,
  options: InsertOptions,
  atStart: boolean,
): { insert: string; prefixOffset: number; textLen: number } {
  const { prefix, suffix = '', block = false } = options;
  if (block) {
    const before = atStart ? '' : '\n\n';
    const after = '\n\n';
    return {
      insert: `${before}${prefix}${text}${suffix}${after}`,
      prefixOffset: before.length + prefix.length,
      textLen: text.length,
    };
  }
  return {
    insert: `${prefix}${text}${suffix}`,
    prefixOffset: prefix.length,
    textLen: text.length,
  };
}


/* =============================================================================
   MarkdownEditor コンポーネント
   ============================================================================= */

export default function MarkdownEditor({ value, onChange, height = '100%' }: MarkdownEditorProps) {
  const { t } = useTranslation();
  const [viewMode, setViewMode] = useState<ViewMode>('split');
  const editorViewRef = useRef<EditorView | null>(null);

  // CodeMirror インスタンスの参照を保持
  const handleEditorCreate = useCallback((view: EditorView) => {
    editorViewRef.current = view;
  }, []);

  // preview モードへ切り替わると CodeMirror がアンマウントされるため参照をクリア
  useEffect(() => {
    if (viewMode === 'preview') {
      editorViewRef.current = null;
    }
  }, [viewMode]);

  // ツールバーから挿入を実行
  const insertMarkdown = useCallback(
    (options: InsertOptions) => {
      const view = editorViewRef.current;
      if (!view || !view.dom.isConnected) {
        // プレビューモード時はエディタが非表示のため split モードへ切り替える
        setViewMode('split');
        return;
      }

      const { defaultText = '' } = options;
      const state = view.state;
      const range = state.selection.main;
      const selected = state.doc.sliceString(range.from, range.to);
      const text = selected || defaultText;
      const atStart = range.from === 0;

      const { insert, prefixOffset, textLen } = buildInsertString(text, options, atStart);
      view.dispatch({
        changes: { from: range.from, to: range.to, insert },
        selection: EditorSelection.range(
          range.from + prefixOffset,
          range.from + prefixOffset + textLen,
        ),
        userEvent: 'input',
      });
      view.focus();
    },
    [value, onChange],
  );

  // キーマップからの stale closure を防ぐため ref 経由で最新の insertMarkdown を参照
  const insertMarkdownRef = useRef(insertMarkdown);
  insertMarkdownRef.current = insertMarkdown;

  const shortcutKeymap = useMemo(
    () =>
      Prec.high(
        keymap.of([
          { key: 'Mod-b', run: () => { insertMarkdownRef.current({ prefix: '**', suffix: '**', defaultText: '太字テキスト' }); return true; } },
          { key: 'Mod-i', run: () => { insertMarkdownRef.current({ prefix: '*', suffix: '*', defaultText: '下線テキスト' }); return true; } },
          { key: 'Mod-k', run: () => { insertMarkdownRef.current({ prefix: '[', suffix: '](URL)', defaultText: 'リンクテキスト' }); return true; } },
          { key: 'Mod-e', run: () => { insertMarkdownRef.current({ prefix: '`', suffix: '`', defaultText: 'コード' }); return true; } },
          { key: 'Mod-Shift-b', run: () => { insertMarkdownRef.current({ prefix: '> ', defaultText: '引用テキスト', block: true }); return true; } },
          { key: 'Mod-Shift-7', run: () => { insertMarkdownRef.current({ prefix: '1. ', defaultText: 'リスト項目', block: true }); return true; } },
          { key: 'Mod-Shift-8', run: () => { insertMarkdownRef.current({ prefix: '- ', defaultText: 'リスト項目', block: true }); return true; } },
          { key: 'Mod-Shift-.', run: () => { insertMarkdownRef.current({ prefix: '- [ ] ', defaultText: 'タスク', block: true }); return true; } },
          { key: 'Mod-Shift-k', run: () => { insertMarkdownRef.current({ prefix: '```\n', suffix: '\n```', defaultText: 'コード', block: true }); return true; } },
          { key: 'Mod-Shift-m', run: () => { insertMarkdownRef.current({ prefix: '$', suffix: '$', defaultText: '数式' }); return true; } },
        ]),
      ),
    [],
  );

  const handleBold = () => insertMarkdown({ prefix: '**', suffix: '**', defaultText: '太字テキスト' });
  const handleItalic = () => insertMarkdown({ prefix: '*', suffix: '*', defaultText: '下線テキスト' });
  const handleStrikethrough = () => insertMarkdown({ prefix: '~~', suffix: '~~', defaultText: '取り消し線' });
  const handleHeading1 = () => insertMarkdown({ prefix: '# ', defaultText: '見出し1', block: true });
  const handleHeading2 = () => insertMarkdown({ prefix: '## ', defaultText: '見出し2', block: true });
  const handleHeading3 = () => insertMarkdown({ prefix: '### ', defaultText: '見出し3', block: true });
  const handleQuote = () => insertMarkdown({ prefix: '> ', defaultText: '引用テキスト', block: true });
  const handleBulletList = () => insertMarkdown({ prefix: '- ', defaultText: 'リスト項目', block: true });
  const handleOrderedList = () => insertMarkdown({ prefix: '1. ', defaultText: 'リスト項目', block: true });
  const handleInlineCode = () => insertMarkdown({ prefix: '`', suffix: '`', defaultText: 'コード' });
  const handleCodeBlock = () =>
    insertMarkdown({ prefix: '```\n', suffix: '\n```', defaultText: 'コード', block: true });
  const handleMath = () => insertMarkdown({ prefix: '$', suffix: '$', defaultText: '数式' });
  const handleLink = () => insertMarkdown({ prefix: '[', suffix: '](URL)', defaultText: 'リンクテキスト' });
  const handleImage = () => insertMarkdown({ prefix: '![', suffix: '](URL)', defaultText: '代替テキスト' });
  const handleTable = () =>
    insertMarkdown({
      prefix: '| 列1 | 列2 | 列3 |\n| --- | --- | --- |\n| セル | セル | セル |',
      block: true,
      defaultText: '',
    });
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
        {/* 見出し */}
        <div className="md-editor-toolbar-group">
          {[
            { onClick: handleHeading1, label: t('editor.toolbar.heading1'), icon: 'H1' },
            { onClick: handleHeading2, label: t('editor.toolbar.heading2'), icon: 'H2' },
            { onClick: handleHeading3, label: t('editor.toolbar.heading3'), icon: 'H3' },
          ].map(({ onClick, label, icon }) => (
            <button
              key={icon}
              type="button"
              className="md-editor-toolbar-btn"
              onClick={onClick}
              title={label}
              aria-label={label}
            >
              <strong style={{ fontSize: '0.75em' }}>{icon}</strong>
            </button>
          ))}
        </div>

        <div className="md-editor-toolbar-separator" aria-hidden="true" />

        {/* インライン装飾 */}
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
            <span style={{ textDecoration: 'underline' }}>I</span>
          </button>
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleStrikethrough}
            title={t('editor.toolbar.strikethrough')}
            aria-label={t('editor.toolbar.strikethrough')}
          >
            <span style={{ textDecoration: 'line-through' }}>S</span>
          </button>
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleInlineCode}
            title={t('editor.toolbar.inlineCode')}
            aria-label={t('editor.toolbar.inlineCode')}
          >
            {'`c`'}
          </button>
        </div>

        <div className="md-editor-toolbar-separator" aria-hidden="true" />

        {/* ブロック要素 */}
        <div className="md-editor-toolbar-group">
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleQuote}
            title={t('editor.toolbar.quote')}
            aria-label={t('editor.toolbar.quote')}
          >
            ❝
          </button>
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleBulletList}
            title={t('editor.toolbar.bulletList')}
            aria-label={t('editor.toolbar.bulletList')}
          >
            ≡
          </button>
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleOrderedList}
            title={t('editor.toolbar.orderedList')}
            aria-label={t('editor.toolbar.orderedList')}
          >
            1≡
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
            onClick={handleMath}
            title={t('editor.toolbar.math')}
            aria-label={t('editor.toolbar.math')}
          >
            ∑
          </button>
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleTable}
            title={t('editor.toolbar.table')}
            aria-label={t('editor.toolbar.table')}
          >
            ⊞
          </button>
        </div>

        <div className="md-editor-toolbar-separator" aria-hidden="true" />

        {/* 挿入 */}
        <div className="md-editor-toolbar-group">
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleLink}
            title={t('editor.toolbar.link')}
            aria-label={t('editor.toolbar.link')}
          >
            🔗
          </button>
          <button
            type="button"
            className="md-editor-toolbar-btn"
            onClick={handleImage}
            title={t('editor.toolbar.image')}
            aria-label={t('editor.toolbar.image')}
          >
            🖼
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
              extensions={[
                shortcutKeymap,
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
            <div className="md-editor-preview-content">
              <MarkdownPreviewContent
                value={value}
                emptyMessage={t('editor.preview.empty')}
              />
            </div>
          </div>
        )}
      </div>

      {/* モバイル: プレビュータブ（スプリットモード時は下にプレビュー） */}
    </div>
  );
}
