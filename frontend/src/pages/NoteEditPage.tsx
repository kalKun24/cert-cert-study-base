import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import MDEditor from '@uiw/react-md-editor';
import rehypeSanitize from 'rehype-sanitize';
import { fetchNote, updateNote } from '../utils/noteApi';
import { fetchTags } from '../utils/tagApi';
import { useAuth } from '../context/AuthContext';
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

const TAB_LABEL_KEYS: Record<EditorTab, string> = {
  body: 'note.bodyLabel',
  discussion_points: 'note.discussionPointsLabel',
  memo: 'note.memoLabel',
};

export default function NoteEditPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { activeTeam } = useTeam();

  const [form, setForm] = useState<FormValues | null>(null);
  const [noteTeamId, setNoteTeamId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<EditorTab>('body');
  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');
  const [validationErrors, setValidationErrors] = useState<
    Partial<Record<keyof FormValues, string>>
  >({});

  useEffect(() => {
    if (!id || !activeTeam) return;
    let isMounted = true;

    Promise.all([fetchNote(activeTeam.id, id), fetchTags(activeTeam.id)])
      .then(([note, tagList]) => {
        if (!isMounted) return;
        // 権限チェック: 作成者本人または admin のみ編集可
        if (user === null || (user.id !== note.created_by && user.role !== 'admin')) {
          navigate(`/notes/${id}`, { replace: true });
          return;
        }
        setNoteTeamId(note.team_id);
        setForm({
          title: note.title,
          body: note.body,
          discussion_points: note.discussion_points,
          memo: note.memo,
          status: note.status,
          selectedTagIds: note.tags,
        });
        setTags(tagList);
      })
      .catch(() => {
        if (isMounted) setLoadError(t('note.error.fetchFailed'));
      })
      .finally(() => {
        if (isMounted) setIsLoading(false);
      });

    return () => {
      isMounted = false;
    };
  }, [id, t, user, navigate, activeTeam]);

  const handleTagToggle = (tagId: string) => {
    if (!form) return;
    setForm((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        selectedTagIds: prev.selectedTagIds.includes(tagId)
          ? prev.selectedTagIds.filter((id) => id !== tagId)
          : [...prev.selectedTagIds, tagId],
      };
    });
  };

  const handleEditorChange = (value: string | undefined) => {
    if (!form) return;
    setForm((prev) => {
      if (!prev) return prev;
      return { ...prev, [activeTab]: value ?? '' };
    });
  };

  const validate = (): boolean => {
    if (!form) return false;
    const errors: Partial<Record<keyof FormValues, string>> = {};
    if (!form.title.trim()) {
      errors.title = t('note.validation.titleRequired');
    }
    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form || !id || !noteTeamId) return;
    if (!validate()) return;

    setIsSubmitting(true);
    setSubmitError('');
    try {
      await updateNote(noteTeamId, id, {
        title: form.title,
        body: form.body,
        discussion_points: form.discussion_points,
        memo: form.memo,
        tags: form.selectedTagIds,
        status: form.status,
      });
      navigate(`/notes/${id}`);
    } catch {
      setSubmitError(t('note.error.updateFailed'));
      setIsSubmitting(false);
    }
  };

  if (isLoading) {
    return (
      <p role="status" className="page-loading">
        {t('common.loading')}
      </p>
    );
  }

  if (loadError || !form) {
    return (
      <div className="page-container-narrow">
        <p role="alert" className="alert alert-error">
          {loadError || t('errors.notFound')}
        </p>
        <Link to={id ? `/notes/${id}` : '/notes'} className="btn btn-secondary">
          {t('note.backToList')}
        </Link>
      </div>
    );
  }

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
              onChange={(e) =>
                setForm((prev) => (prev ? { ...prev, title: e.target.value } : prev))
              }
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
              setForm((prev) =>
                prev ? { ...prev, status: e.target.value as NoteStatus } : prev
              )
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
              onClick={() => navigate(`/notes/${id}`)}
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

        {/* MDEditor（全幅・全高さ） */}
        <div
          className="editor-wrapper"
          data-color-mode="light"
          role="tabpanel"
          id={`tabpanel-${activeTab}`}
          aria-labelledby={`tab-${activeTab}`}
        >
          <MDEditor
            value={form[activeTab]}
            onChange={handleEditorChange}
            height="100%"
            preview="live"
            previewOptions={{
              rehypePlugins: [[rehypeSanitize]],
            }}
          />
        </div>
      </div>
    </div>
  );
}
