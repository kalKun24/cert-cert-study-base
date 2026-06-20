import { useState, useEffect, FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../context/AuthContext';
import { useTeam } from '../context/TeamContext';
import { Tag } from '../types/tag';
import { fetchTags, createTag, updateTag, deleteTag } from '../utils/tagApi';

export default function TagManagePage() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const { activeTeam } = useTeam();
  const isAdmin = user?.role === 'admin';

  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState('');

  // 作成フォーム
  const [newName, setNewName] = useState('');
  const [createError, setCreateError] = useState('');
  const [isCreating, setIsCreating] = useState(false);

  // 編集状態（admin のみ）
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');
  const [editError, setEditError] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  const loadTags = () => {
    if (!activeTeam) return;
    setIsLoading(true);
    setFetchError('');
    fetchTags(activeTeam.id)
      .then(setTags)
      .catch(() => setFetchError(t('tag.error.fetchFailed')))
      .finally(() => setIsLoading(false));
  };

  useEffect(() => {
    if (activeTeam) {
      loadTags();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTeam?.id]);

  const handleCreate = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!activeTeam) return;
    const trimmed = newName.trim();
    if (!trimmed) {
      setCreateError(t('tag.validation.nameRequired'));
      return;
    }
    setIsCreating(true);
    setCreateError('');
    try {
      const created = await createTag(activeTeam.id, trimmed);
      setTags((prev) => [...prev, created]);
      setNewName('');
    } catch {
      setCreateError(t('tag.error.createFailed'));
    } finally {
      setIsCreating(false);
    }
  };

  const startEdit = (tag: Tag) => {
    setEditingId(tag.id);
    setEditingName(tag.name);
    setEditError('');
  };

  const cancelEdit = () => {
    setEditingId(null);
    setEditingName('');
    setEditError('');
  };

  const handleSaveEdit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!activeTeam || !editingId) return;
    const trimmed = editingName.trim();
    if (!trimmed) {
      setEditError(t('tag.validation.nameRequired'));
      return;
    }
    setIsSaving(true);
    setEditError('');
    try {
      const updated = await updateTag(activeTeam.id, editingId, trimmed);
      setTags((prev) => prev.map((tag) => (tag.id === updated.id ? updated : tag)));
      cancelEdit();
    } catch {
      setEditError(t('tag.error.updateFailed'));
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (tag: Tag) => {
    if (!activeTeam) return;
    const confirmed = window.confirm(t('tag.confirm.delete', { name: tag.name }));
    if (!confirmed) return;
    try {
      await deleteTag(activeTeam.id, tag.id);
      setTags((prev) => prev.filter((t) => t.id !== tag.id));
    } catch {
      window.alert(t('tag.error.deleteFailed'));
    }
  };

  if (!activeTeam) {
    return (
      <section className="tag-manage-page page-container-full">
        <h1 className="page-title">{t('tag.list.title')}</h1>
        <p className="tag-list-empty">{t('tag.list.noTeam')}</p>
      </section>
    );
  }

  return (
    <section className="tag-manage-page page-container-full">
      <h1 className="page-title">{t('tag.list.title')}</h1>

      <div className="tag-create-form-wrapper">
        <h2 className="section-title">{t('tag.form.createTitle')}</h2>
        <form onSubmit={handleCreate} className="tag-create-form" noValidate>
          <label htmlFor="new-tag-name" className="form-label">
            {t('tag.form.nameLabel')}
          </label>
          <div className="tag-create-form-row">
            <input
              id="new-tag-name"
              type="text"
              className="form-input"
              placeholder={t('tag.form.namePlaceholder')}
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              aria-describedby={createError ? 'new-tag-error' : undefined}
            />
            <button type="submit" className="btn btn-primary" disabled={isCreating}>
              {t('tag.form.addButton')}
            </button>
          </div>
          {createError && (
            <p id="new-tag-error" role="alert" className="form-error">
              {createError}
            </p>
          )}
        </form>
      </div>

      {isLoading ? (
        <p role="status" className="page-loading">
          {t('common.loading')}
        </p>
      ) : fetchError ? (
        <p role="alert" className="alert alert-error">
          {fetchError}
        </p>
      ) : tags.length === 0 ? (
        <p className="tag-list-empty">{t('tag.list.empty')}</p>
      ) : (
        <ul className="tag-manage-list">
          {tags.map((tag) => (
            <li key={tag.id} className="tag-manage-item">
              {isAdmin && editingId === tag.id ? (
                <form onSubmit={handleSaveEdit} className="tag-edit-form" noValidate>
                  <label htmlFor={`edit-tag-${tag.id}`} className="sr-only">
                    {t('tag.form.nameLabel')}
                  </label>
                  <div className="tag-edit-form-row">
                    <input
                      id={`edit-tag-${tag.id}`}
                      type="text"
                      className="form-input"
                      value={editingName}
                      onChange={(e) => setEditingName(e.target.value)}
                      aria-describedby={editError ? `edit-tag-error-${tag.id}` : undefined}
                      autoFocus
                    />
                    <button type="submit" className="btn btn-primary" disabled={isSaving}>
                      {t('tag.form.saveButton')}
                    </button>
                    <button
                      type="button"
                      className="btn btn-secondary"
                      onClick={cancelEdit}
                      disabled={isSaving}
                    >
                      {t('common.cancel')}
                    </button>
                  </div>
                  {editError && (
                    <p id={`edit-tag-error-${tag.id}`} role="alert" className="form-error">
                      {editError}
                    </p>
                  )}
                </form>
              ) : (
                <div className="tag-manage-item-row">
                  <span className="tag-manage-name">{tag.name}</span>
                  <div className="tag-manage-actions">
                    {isAdmin && (
                      <button
                        type="button"
                        className="btn btn-secondary btn-sm"
                        onClick={() => startEdit(tag)}
                      >
                        {t('common.edit')}
                      </button>
                    )}
                    <button
                      type="button"
                      className="btn btn-danger btn-sm"
                      onClick={() => handleDelete(tag)}
                    >
                      {t('common.delete')}
                    </button>
                  </div>
                </div>
              )}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
