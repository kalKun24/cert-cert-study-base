import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import MDEditor from '@uiw/react-md-editor';
import { createQuestion } from '../utils/questionApi';
import { fetchTags } from '../utils/tagApi';
import { useTeam } from '../context/TeamContext';
import { QuestionStatus, VisibilityScope } from '../types/question';
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

    setIsSubmitting(true);
    setSubmitError('');
    try {
      const created = await createQuestion({
        title: form.title,
        body: form.body,
        answer: form.answer,
        explanation: form.explanation,
        memo: form.memo,
        tags: form.selectedTagNames,
        status: form.status,
        visibility_scope: 'all' as VisibilityScope,
        published_team_ids: [],
      });
      navigate(`/questions/${created.id}`);
    } catch {
      setSubmitError(t('question.error.createFailed'));
      setIsSubmitting(false);
    }
  };

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
              onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))}
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
            selectedTagNames={form.selectedTagNames}
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
              setForm((prev) => ({ ...prev, status: e.target.value as QuestionStatus }))
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
              onClick={() => navigate('/questions')}
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
              aria-selected={activeTab === tab}
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
        <div className="editor-wrapper" data-color-mode="light">
          <MDEditor
            value={form[activeTab]}
            onChange={handleEditorChange}
            height="100%"
            preview="live"
          />
        </div>
      </div>
    </div>
  );
}
