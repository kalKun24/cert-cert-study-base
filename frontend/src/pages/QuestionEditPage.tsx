import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import MDEditor from '@uiw/react-md-editor';
import rehypeSanitize from 'rehype-sanitize';
import { fetchQuestion, updateQuestion } from '../utils/questionApi';
import { fetchTags } from '../utils/tagApi';
import { useAuth } from '../context/AuthContext';
import { QuestionStatus } from '../types/question';
import { Tag } from '../types/tag';

type EditorTab = 'body' | 'answer' | 'explanation' | 'memo';

const EDITOR_TABS: EditorTab[] = ['body', 'answer', 'explanation', 'memo'];

interface FormValues {
  title: string;
  body: string;
  answer: string;
  explanation: string;
  memo: string;
  status: QuestionStatus;
  selectedTagNames: string[];
}

export default function QuestionEditPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const [form, setForm] = useState<FormValues | null>(null);
  const [activeTab, setActiveTab] = useState<EditorTab>('body');
  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');
  const [validationErrors, setValidationErrors] = useState<Partial<Record<keyof FormValues, string>>>({});

  useEffect(() => {
    if (!id) return;
    let isMounted = true;

    Promise.all([
      fetchQuestion(id),
      fetchTags(),
    ])
      .then(([q, tagList]) => {
        if (!isMounted) return;
        // 権限チェック: 作成者本人または admin のみ編集可
        if (user === null || (user.id !== q.created_by && user.role !== 'admin')) {
          navigate(`/questions/${id}`, { replace: true });
          return;
        }
        setForm({
          title: q.title,
          body: q.body,
          answer: q.answer,
          explanation: q.explanation,
          memo: q.memo,
          status: q.status,
          selectedTagNames: q.tags,
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
  }, [id, t, user, navigate]);

  const handleTagToggle = (tagName: string) => {
    if (!form) return;
    setForm((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        selectedTagNames: prev.selectedTagNames.includes(tagName)
          ? prev.selectedTagNames.filter((n) => n !== tagName)
          : [...prev.selectedTagNames, tagName],
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
    if (!form || !id) return;
    if (!validate()) return;

    setIsSubmitting(true);
    setSubmitError('');
    try {
      await updateQuestion(id, {
        title: form.title,
        body: form.body,
        answer: form.answer,
        explanation: form.explanation,
        memo: form.memo,
        tags: form.selectedTagNames,
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
      <div>
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
    <section className="question-form-page">
      <h1 className="page-title">{t('question.edit')}</h1>

      <form onSubmit={handleSubmit} noValidate className="question-form">
        <div className="form-field">
          <label htmlFor="question-title" className="form-label">
            {t('question.titleLabel')}
          </label>
          <input
            id="question-title"
            type="text"
            className="form-input"
            value={form.title}
            onChange={(e) => setForm((prev) => prev ? { ...prev, title: e.target.value } : prev)}
            aria-describedby={validationErrors.title ? 'title-error' : undefined}
            aria-invalid={!!validationErrors.title}
          />
          {validationErrors.title && (
            <p id="title-error" role="alert" className="form-error">
              {validationErrors.title}
            </p>
          )}
        </div>

        <div className="form-field">
          <div className="editor-tabs" role="tablist">
            {EDITOR_TABS.map((tab) => (
              <button
                key={tab}
                type="button"
                role="tab"
                aria-selected={activeTab === tab}
                className={`editor-tab${activeTab === tab ? ' editor-tab--active' : ''}`}
                onClick={() => setActiveTab(tab)}
              >
                {t(`question.${tab}Label`)}
              </button>
            ))}
          </div>

          <div data-color-mode="light">
            <MDEditor
              value={form[activeTab]}
              onChange={handleEditorChange}
              height={300}
              preview="live"
              previewOptions={{
                rehypePlugins: [[rehypeSanitize]],
              }}
            />
          </div>
          {validationErrors.body && (
            <p role="alert" className="form-error">
              {validationErrors.body}
            </p>
          )}
        </div>

        <div className="form-field">
          <fieldset>
            <legend className="form-label">{t('question.tagSelectLabel')}</legend>
            <div className="tag-checkbox-list">
              {tags.map((tag) => (
                <label key={tag.id} className="tag-checkbox-label">
                  <input
                    type="checkbox"
                    checked={form.selectedTagNames.includes(tag.name)}
                    onChange={() => handleTagToggle(tag.name)}
                  />
                  {tag.name}
                </label>
              ))}
            </div>
          </fieldset>
        </div>

        <div className="form-field">
          <label htmlFor="question-status" className="form-label">
            {t('question.statusLabel')}
          </label>
          <select
            id="question-status"
            className="form-select"
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
        </div>

        {submitError && (
          <p role="alert" className="alert alert-error">
            {submitError}
          </p>
        )}

        <div className="form-actions">
          <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
            {t('common.save')}
          </button>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => navigate(`/questions/${id}`)}
            disabled={isSubmitting}
          >
            {t('common.cancel')}
          </button>
        </div>
      </form>
    </section>
  );
}
