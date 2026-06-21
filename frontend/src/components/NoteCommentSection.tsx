import { useState, useEffect, FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import ReactMarkdown from 'react-markdown';
import rehypeSanitize from 'rehype-sanitize';
import { useAuth } from '../context/AuthContext';
import { NoteComment } from '../types/noteComment';
import {
  fetchNoteComments,
  postNoteComment,
  updateNoteComment,
  deleteNoteComment,
} from '../utils/noteCommentApi';

interface Props {
  teamId: string;
  noteId: string;
}

export default function NoteCommentSection({ teamId, noteId }: Props) {
  const { t } = useTranslation();
  const { user } = useAuth();

  const [comments, setComments] = useState<NoteComment[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');

  // 新規投稿フォームの状態
  const [newBody, setNewBody] = useState('');
  const [newPreview, setNewPreview] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');

  // 削除エラー
  const [deleteError, setDeleteError] = useState('');

  // 編集中のコメント管理
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editBody, setEditBody] = useState('');
  const [editPreview, setEditPreview] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [editError, setEditError] = useState('');

  useEffect(() => {
    let isMounted = true;
    setIsLoading(true);
    setLoadError('');
    fetchNoteComments(teamId, noteId)
      .then((data) => {
        if (isMounted) setComments(data);
      })
      .catch(() => {
        if (isMounted) setLoadError(t('comment.error.fetchFailed'));
      })
      .finally(() => {
        if (isMounted) setIsLoading(false);
      });
    return () => {
      isMounted = false;
    };
  }, [teamId, noteId, t]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!newBody.trim()) {
      setSubmitError(t('comment.validation.bodyRequired'));
      return;
    }
    setIsSubmitting(true);
    setSubmitError('');
    try {
      const created = await postNoteComment(teamId, noteId, newBody.trim());
      setComments((prev) => [...prev, created]);
      setNewBody('');
      setNewPreview(false);
    } catch {
      setSubmitError(t('comment.error.postFailed'));
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleEditStart = (comment: NoteComment) => {
    setEditingId(comment.id);
    setEditBody(comment.body);
    setEditPreview(false);
    setEditError('');
  };

  const handleEditCancel = () => {
    setEditingId(null);
    setEditBody('');
    setEditPreview(false);
    setEditError('');
  };

  const handleEditSave = async (commentId: string) => {
    if (!editBody.trim()) {
      setEditError(t('comment.validation.bodyRequired'));
      return;
    }
    setIsUpdating(true);
    setEditError('');
    try {
      const updated = await updateNoteComment(teamId, noteId, commentId, editBody.trim());
      setComments((prev) => prev.map((c) => (c.id === commentId ? updated : c)));
      setEditingId(null);
      setEditBody('');
    } catch {
      setEditError(t('comment.error.updateFailed'));
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDelete = async (commentId: string) => {
    if (!window.confirm(t('comment.confirm.delete'))) return;
    setDeleteError('');
    try {
      await deleteNoteComment(teamId, noteId, commentId);
      setComments((prev) => prev.filter((c) => c.id !== commentId));
    } catch {
      setDeleteError(t('comment.error.deleteFailed'));
    }
  };

  const formatDate = (iso: string): string => {
    const d = new Date(iso);
    return d.toLocaleString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <section className="comment-section" aria-label={t('comment.section.title')}>
      <h2 className="comment-section-title">{t('comment.section.title')}</h2>

      {/* コメント一覧 */}
      {isLoading && (
        <p role="status" aria-live="polite" className="comment-loading">
          {t('common.loading')}
        </p>
      )}
      {loadError && (
        <p role="alert" className="alert alert-error">
          {loadError}
        </p>
      )}
      {!isLoading && !loadError && comments.length === 0 && (
        <p className="comment-empty">{t('comment.section.empty')}</p>
      )}

      {deleteError && (
        <p role="alert" className="alert alert-error">
          {deleteError}
        </p>
      )}

      {!isLoading && !loadError && comments.length > 0 && (
        <ul className="comment-list" aria-label={t('comment.section.title')}>
          {comments.map((comment) => {
            const isOwner = user?.id === comment.created_by;
            const isAdmin = user?.role === 'admin';
            const canDelete = isOwner || isAdmin;
            const canEdit = isOwner;
            const isEditing = editingId === comment.id;

            return (
              <li
                key={comment.id}
                className={`comment-item${isOwner ? ' comment-item--own' : ''}`}
              >
                <div className="comment-header">
                  <span className="comment-author">
                    {comment.display_name}
                    {isOwner && (
                      <span className="comment-own-badge">{t('comment.badge.own')}</span>
                    )}
                  </span>
                  <span className="comment-date">{formatDate(comment.created_at)}</span>
                  {(canEdit || canDelete) && !isEditing && (
                    <div
                      className="comment-actions"
                      role="group"
                      aria-label={t('comment.section.title')}
                    >
                      {canEdit && (
                        <button
                          type="button"
                          className="btn btn-secondary btn-sm"
                          onClick={() => handleEditStart(comment)}
                          aria-label={`${t('common.edit')} - ${comment.display_name}のコメント`}
                        >
                          {t('common.edit')}
                        </button>
                      )}
                      {canDelete && (
                        <button
                          type="button"
                          className="btn btn-danger btn-sm"
                          onClick={() => handleDelete(comment.id)}
                          aria-label={`${t('common.delete')} - ${comment.display_name}のコメント`}
                        >
                          {t('common.delete')}
                        </button>
                      )}
                    </div>
                  )}
                </div>

                {/* 編集モード */}
                {isEditing ? (
                  <div className="comment-edit-form">
                    <div className="comment-preview-toggle">
                      <button
                        type="button"
                        className={`btn btn-tab${!editPreview ? ' btn-tab--active' : ''}`}
                        onClick={() => setEditPreview(false)}
                      >
                        {t('comment.form.write')}
                      </button>
                      <button
                        type="button"
                        className={`btn btn-tab${editPreview ? ' btn-tab--active' : ''}`}
                        onClick={() => setEditPreview(true)}
                      >
                        {t('comment.form.preview')}
                      </button>
                    </div>

                    {editPreview ? (
                      <div className="comment-preview-body">
                        {editBody ? (
                          <ReactMarkdown rehypePlugins={[rehypeSanitize]}>
                            {editBody}
                          </ReactMarkdown>
                        ) : (
                          <span className="comment-preview-empty">
                            {t('comment.form.previewEmpty')}
                          </span>
                        )}
                      </div>
                    ) : (
                      <textarea
                        className="comment-textarea"
                        value={editBody}
                        onChange={(e) => setEditBody(e.target.value)}
                        rows={4}
                        disabled={isUpdating}
                        aria-label={t('comment.form.bodyLabel')}
                      />
                    )}

                    {editError && (
                      <p role="alert" className="alert alert-error">
                        {editError}
                      </p>
                    )}

                    <div className="comment-edit-actions">
                      <button
                        type="button"
                        className="btn btn-primary"
                        onClick={() => handleEditSave(comment.id)}
                        disabled={isUpdating}
                      >
                        {isUpdating ? t('common.loading') : t('common.save')}
                      </button>
                      <button
                        type="button"
                        className="btn btn-secondary"
                        onClick={handleEditCancel}
                        disabled={isUpdating}
                      >
                        {t('common.cancel')}
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="comment-body">
                    <ReactMarkdown rehypePlugins={[rehypeSanitize]}>
                      {comment.body}
                    </ReactMarkdown>
                  </div>
                )}
              </li>
            );
          })}
        </ul>
      )}

      {/* 新規コメント投稿フォーム */}
      <div className="comment-new-form">
        <h3 className="comment-new-form-title">{t('comment.form.title')}</h3>
        <form onSubmit={handleSubmit} noValidate>
          <div className="comment-preview-toggle">
            <button
              type="button"
              className={`btn btn-tab${!newPreview ? ' btn-tab--active' : ''}`}
              onClick={() => setNewPreview(false)}
            >
              {t('comment.form.write')}
            </button>
            <button
              type="button"
              className={`btn btn-tab${newPreview ? ' btn-tab--active' : ''}`}
              onClick={() => setNewPreview(true)}
            >
              {t('comment.form.preview')}
            </button>
          </div>

          {newPreview ? (
            <div className="comment-preview-body">
              {newBody ? (
                <ReactMarkdown rehypePlugins={[rehypeSanitize]}>{newBody}</ReactMarkdown>
              ) : (
                <span className="comment-preview-empty">{t('comment.form.previewEmpty')}</span>
              )}
            </div>
          ) : (
            <textarea
              className="comment-textarea"
              value={newBody}
              onChange={(e) => setNewBody(e.target.value)}
              placeholder={t('comment.form.placeholder')}
              rows={4}
              disabled={isSubmitting}
              aria-label={t('comment.form.bodyLabel')}
            />
          )}

          {submitError && (
            <p role="alert" className="alert alert-error">
              {submitError}
            </p>
          )}

          <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
            {isSubmitting ? t('common.loading') : t('comment.form.submit')}
          </button>
        </form>
      </div>
    </section>
  );
}
