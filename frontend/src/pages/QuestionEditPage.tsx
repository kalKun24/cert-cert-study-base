import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import MDEditor from '@uiw/react-md-editor';
import rehypeSanitize from 'rehype-sanitize';
import { fetchQuestion, updateQuestion } from '../utils/questionApi';
import { fetchTags } from '../utils/tagApi';
import { useAuth } from '../context/AuthContext';
import { useTeam } from '../context/TeamContext';
import { QuestionStatus } from '../types/question';
import { Tag } from '../types/tag';
import TagDropdown from '../components/TagDropdown';

type EditorTab = 'body' | 'answer' | 'explanation' | 'memo';

const EDITOR_TABS: EditorTab[] = ['body', 'answer', 'explanation', 'memo'];

interface FormValues {
  title: string;
  body: string;
  answer: string;
  explanation: string;
  memo: string;
  status: QuestionStatus;
  selectedTagIds: string[];
}

export default function QuestionEditPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { activeTeam } = useTeam();

  const [form, setForm] = useState<FormValues | null>(null);
  const [questionTeamId, setQuestionTeamId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<EditorTab>('body');
  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');
  const [validationErrors, setValidationErrors] = useState<Partial<Record<keyof FormValues, string>>>({});

  useEffect(() => {
    if (!id || !activeTeam) return;
    let isMounted = true;

    Promise.all([
      fetchQuestion(activeTeam.id, id),
      fetchTags(activeTeam.id),
    ])
      .then(([q, tagList]) => {
        if (!isMounted) return;
        // 権限チェック: 作成者本人または admin のみ編集可
        if (user === null || (user.id !== q.created_by && user.role !== 'admin')) {
          navigate(`/questions/${id}`, { replace: true });
          return;
        }
        setQuestionTeamId(q.team_id);
        setForm({
          title: q.title,
          body: q.body,
          answer: q.answer,
          explanation: q.explanation,
          memo: q.memo,
          status: q.status,
          selectedTagIds: q.tags,
        });
        setTags(tagList);
      })
      .catch(() => {
        if (isMounted) setLoadError(t('question.error.fetchFailed'));
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
      errors.title = t('question.validation.titleRequired');
    }
    if (!form.body.trim()) {
      errors.body = t('question.validation.bodyRequired');
    }
    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form || !id || !questionTeamId) return;
    if (!validate()) return;

    setIsSubmitting(true);
    setSubmitError('');
    try {
      await updateQuestion(questionTeamId, id, {
        title: form.title,
        body: form.body,
        answer: form.answer,
        explanation: form.explanation,
        memo: form.memo,
        tags: form.selectedTagIds,
        status: form.status,
      });
      navigate(`/questions/${id}`);
    } catch {
      setSubmitError(t('question.error.updateFailed'));
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
        <Link to={id ? `/questions/${id}` : '/questions'} className="btn btn-secondary">
          {t('question.backToList')}
        </Link>
      </div>
    );
  }

  return (
    // エディタページはレイアウト全体を占有するため、通常のコンテナを使わない
    <div className="editor-page">
      {/* ===== メタバー（sticky） ===== */}
      <form onSubmit={handleSubmit} noValidate className="editor-meta-bar">
        <div className="editor-meta-bar-inner">
          {/* タイトル入力（flex:1 で横幅を最大化） */}
          <div className="editor-meta-title-field">
            <label htmlFor="question-title" className="sr-only">
              {t('question.titleLabel')}
            </label>
            <input
              id="question-title"
              type="text"
              className={`editor-meta-title-input${validationErrors.title ? ' editor-meta-title-input--error' : ''}`}
              placeholder={t('question.titleLabel')}
              value={form.title}
              onChange={(e) => setForm((prev) => prev ? { ...prev, title: e.target.value } : prev)}
              aria-describedby={validationErrors.title ? 'title-error' : undefined}
              aria-invalid={!!validationErrors.title}
            />
            {validationErrors.title && (
              <p id="title-error" role="alert" className="editor-meta-field-error">
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
          <label htmlFor="question-status" className="sr-only">
            {t('question.statusLabel')}
          </label>
          <select
            id="question-status"
            className="editor-meta-select"
            value={form.status}
            onChange={(e) =>
              setForm((prev) =>
                prev ? { ...prev, status: e.target.value as QuestionStatus } : prev
              )
            }
          >
            <option value="draft">{t('question.status.draft')}</option>
            <option value="private">{t('question.status.private')}</option>
            <option value="published">{t('question.status.published')}</option>
          </select>

          {/* アクションボタン */}
          <div className="editor-meta-actions">
            <button type="submit" className="btn btn-primary btn-sm" disabled={isSubmitting}>
              {t('common.save')}
            </button>
            <button
              type="button"
              className="btn btn-secondary btn-sm"
              onClick={() => navigate(`/questions/${id}`)}
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
              {t(`question.${tab}Label`)}
            </button>
          ))}
        </div>

        {/* バリデーションエラー（問題文タブ選択時のみ） */}
        {activeTab === 'body' && validationErrors.body && (
          <p role="alert" className="editor-body-error">
            {validationErrors.body}
          </p>
        )}

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
