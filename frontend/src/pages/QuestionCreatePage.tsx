import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import MDEditor from '@uiw/react-md-editor';
import { createQuestion } from '../utils/questionApi';
import { fetchTags } from '../utils/tagApi';
import { useTeam } from '../context/TeamContext';
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

const INITIAL_FORM: FormValues = {
  title: '',
  body: '',
  answer: '',
  explanation: '',
  memo: '',
  status: 'draft',
  selectedTagNames: [],
};

export default function QuestionCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { activeTeam } = useTeam();

  const [form, setForm] = useState<FormValues>(INITIAL_FORM);
  const [activeTab, setActiveTab] = useState<EditorTab>('body');
  const [tags, setTags] = useState<Tag[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');
  const [validationErrors, setValidationErrors] = useState<Partial<Record<keyof FormValues, string>>>({});

  useEffect(() => {
    if (!activeTeam) return;
    fetchTags(activeTeam.id).then(setTags).catch(() => {
      // タグ取得エラーは非致命的
    });
  }, [activeTeam?.id]);

  const handleTagToggle = (tagName: string) => {
    setForm((prev) => ({
      ...prev,
      selectedTagNames: prev.selectedTagNames.includes(tagName)
        ? prev.selectedTagNames.filter((n) => n !== tagName)
        : [...prev.selectedTagNames, tagName],
    }));
  };

  const handleEditorChange = (value: string | undefined) => {
    setForm((prev) => ({ ...prev, [activeTab]: value ?? '' }));
  };

  const validate = (): boolean => {
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
    if (!validate()) return;
    if (!activeTeam) return;

    setIsSubmitting(true);
    setSubmitError('');
    try {
      const created = await createQuestion(activeTeam.id, {
        title: form.title,
        body: form.body,
        answer: form.answer,
        explanation: form.explanation,
        memo: form.memo,
        tags: form.selectedTagNames,
        status: form.status,
      });
      navigate(`/questions/${created.id}`);
    } catch {
      setSubmitError(t('question.error.createFailed'));
      setIsSubmitting(false);
    }
  };

  return (
    <section className="question-form-page content-narrow page-container-narrow">
      <h1 className="page-title">{t('question.new')}</h1>

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
            onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))}
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
            />
          </div>
          {activeTab === 'body' && validationErrors.body && (
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
              setForm((prev) => ({ ...prev, status: e.target.value as QuestionStatus }))
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
            onClick={() => navigate('/questions')}
            disabled={isSubmitting}
          >
            {t('common.cancel')}
          </button>
        </div>
      </form>
    </section>
  );
}
