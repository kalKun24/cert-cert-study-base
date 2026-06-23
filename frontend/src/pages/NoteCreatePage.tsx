import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import MarkdownEditor from '../components/MarkdownEditor';
import { createNote } from '../utils/noteApi';
import { fetchTags } from '../utils/tagApi';
import { useTeam } from '../context/TeamContext';
import { NoteStatus } from '../types/note';
import { Tag } from '../types/tag';
import TagDropdown from '../components/TagDropdown';

type EditorTab = 'body' | 'discussion_points' | 'memo';

const EDITOR_TABS: EditorTab[] = ['body', 'discussion_points', 'memo'];

interface FormValues {
  title: string;
  body: string;
  discussion_points: string;
  memo: string;
  status: NoteStatus;
  selectedTagIds: string[];
}

const INITIAL_FORM: FormValues = {
  title: '',
  body: '',
  discussion_points: '',
  memo: '',
  status: 'draft',
  selectedTagIds: [],
};

const TAB_LABEL_KEYS: Record<EditorTab, string> = {
  body: 'note.bodyLabel',
  discussion_points: 'note.discussionPointsLabel',
  memo: 'note.memoLabel',
};

export default function NoteCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { activeTeam } = useTeam();

  const [form, setForm] = useState<FormValues>(INITIAL_FORM);
  const [activeTab, setActiveTab] = useState<EditorTab>('body');
  const [tags, setTags] = useState<Tag[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');
  const [validationErrors, setValidationErrors] = useState<
    Partial<Record<keyof FormValues, string>>
  >({});

  useEffect(() => {
    if (!activeTeam) return;
    fetchTags(activeTeam.id)
      .then(setTags)
      .catch(() => {
        // タグ取得エラーは非致命的
      });
  }, [activeTeam?.id]);

  const handleTagToggle = (tagId: string) => {
    setForm((prev) => ({
      ...prev,
      selectedTagIds: prev.selectedTagIds.includes(tagId)
        ? prev.selectedTagIds.filter((id) => id !== tagId)
        : [...prev.selectedTagIds, tagId],
    }));
  };

  const handleEditorChange = (value: string) => {
    setForm((prev) => ({ ...prev, [activeTab]: value }));
  };

  const validate = (): boolean => {
    const errors: Partial<Record<keyof FormValues, string>> = {};
    if (!form.title.trim()) {
      errors.title = t('note.validation.titleRequired');
    }
    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;
    if (!activeTeam) return;

    setIsSubmitting(true);
    setSubmitError('');
    try {
      const created = await createNote(activeTeam.id, {
        title: form.title,
        body: form.body,
        discussion_points: form.discussion_points,
        memo: form.memo,
        tags: form.selectedTagIds,
        status: form.status,
      });
      navigate(`/notes/${created.id}`);
    } catch {
      setSubmitError(t('note.error.createFailed'));
      setIsSubmitting(false);
    }
  };

  return (
    <div className="editor-page">
      {/* ===== メタバー（sticky） ===== */}
      <form onSubmit={handleSubmit} noValidate className="editor-meta-bar">
        <div className="editor-meta-bar-inner">
          {/* タイトル入力 */}
          <div className="editor-meta-title-field">
            <label htmlFor="note-title" className="sr-only">
              {t('note.titleLabel')}
            </label>
            <input
              id="note-title"
              type="text"
              className={`editor-meta-title-input${validationErrors.title ? ' editor-meta-title-input--error' : ''}`}
              placeholder={t('note.titleLabel')}
              value={form.title}
              onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))}
              aria-describedby={validationErrors.title ? 'note-title-error' : undefined}
              aria-invalid={!!validationErrors.title}
            />
            {validationErrors.title && (
              <p id="note-title-error" role="alert" className="editor-meta-field-error">
                {validationErrors.title}
              </p>
            )}
          </div>

          {/* タグドロップダウン */}
          <TagDropdown
            tags={tags}
            selectedTagIds={form.selectedTagIds}
            onToggle={handleTagToggle}
          />

          {/* ステータス選択 */}
          <label htmlFor="note-status" className="sr-only">
            {t('note.statusLabel')}
          </label>
          <select
            id="note-status"
            className="editor-meta-select"
            value={form.status}
            onChange={(e) =>
              setForm((prev) => ({ ...prev, status: e.target.value as NoteStatus }))
            }
          >
            <option value="draft">{t('note.status.draft')}</option>
            <option value="private">{t('note.status.private')}</option>
            <option value="published">{t('note.status.published')}</option>
          </select>

          {/* アクションボタン */}
          <div className="editor-meta-actions">
            <button type="submit" className="btn btn-primary btn-sm" disabled={isSubmitting}>
              {t('common.save')}
            </button>
            <button
              type="button"
              className="btn btn-secondary btn-sm"
              onClick={() => navigate('/notes')}
              disabled={isSubmitting}
            >
              {t('common.cancel')}
            </button>
          </div>
        </div>

        {/* 送信エラー */}
        {submitError && (
          <p role="alert" className="editor-meta-submit-error">
            {submitError}
          </p>
        )}
      </form>

      {/* ===== エディタ本体（タブ + MDEditor） ===== */}
      <div className="editor-body">
        {/* タブ切り替え */}
        <div className="editor-tabs editor-tabs--fullwidth" role="tablist">
          {EDITOR_TABS.map((tab) => (
            <button
              key={tab}
              type="button"
              role="tab"
              id={`tab-${tab}`}
              aria-selected={activeTab === tab}
              aria-controls={`tabpanel-${tab}`}
              className={`editor-tab${activeTab === tab ? ' editor-tab--active' : ''}`}
              onClick={() => setActiveTab(tab)}
            >
              {t(TAB_LABEL_KEYS[tab])}
            </button>
          ))}
        </div>

        {/* MarkdownEditor（全幅・全高さ） */}
        <div
          className="editor-wrapper"
          role="tabpanel"
          id={`tabpanel-${activeTab}`}
          aria-labelledby={`tab-${activeTab}`}
        >
          <MarkdownEditor
            value={form[activeTab]}
            onChange={handleEditorChange}
            height="100%"
          />
        </div>
      </div>
    </div>
  );
}
