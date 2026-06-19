import { useState } from 'react';
import { useTranslation } from 'react-i18next';

interface TeamFormProps {
  initialName?: string;
  initialDescription?: string;
  isSubmitting: boolean;
  submitError: string;
  onSubmit: (name: string, description: string) => void;
  onCancel: () => void;
}

interface ValidationErrors {
  name?: string;
}

export default function TeamForm({
  initialName = '',
  initialDescription = '',
  isSubmitting,
  submitError,
  onSubmit,
  onCancel,
}: TeamFormProps) {
  const { t } = useTranslation();

  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDescription);
  const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});

  const validate = (): boolean => {
    const errors: ValidationErrors = {};
    if (!name.trim()) {
      errors.name = t('team.validation.nameRequired');
    }
    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;
    onSubmit(name.trim(), description.trim());
  };

  return (
    <form onSubmit={handleSubmit} noValidate className="team-form">
      <div className="form-field">
        <label htmlFor="team-name" className="form-label">
          {t('team.form.nameLabel')}
        </label>
        <input
          id="team-name"
          type="text"
          className="form-input"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder={t('team.form.namePlaceholder')}
          aria-describedby={validationErrors.name ? 'team-name-error' : undefined}
          aria-invalid={!!validationErrors.name}
          disabled={isSubmitting}
        />
        {validationErrors.name && (
          <p id="team-name-error" role="alert" className="form-error">
            {validationErrors.name}
          </p>
        )}
      </div>

      <div className="form-field">
        <label htmlFor="team-description" className="form-label">
          {t('team.form.descriptionLabel')}
        </label>
        <textarea
          id="team-description"
          className="form-textarea"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder={t('team.form.descriptionPlaceholder')}
          rows={4}
          disabled={isSubmitting}
        />
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
          onClick={onCancel}
          disabled={isSubmitting}
        >
          {t('common.cancel')}
        </button>
      </div>
    </form>
  );
}
