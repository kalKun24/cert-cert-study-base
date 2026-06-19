import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { UserRole, CreateUserRequest, UpdateUserRequest } from '../types/user';

const USER_ROLES: UserRole[] = ['admin', 'teamowner', 'user'];

interface UserFormProps {
  mode: 'create' | 'edit';
  initialDisplayName?: string;
  initialEmail?: string;
  initialRole?: UserRole;
  initialUsername?: string;
  isSubmitting: boolean;
  submitError: string;
  onSubmit: (data: CreateUserRequest | UpdateUserRequest) => void;
  onCancel: () => void;
}

interface ValidationErrors {
  username?: string;
  displayName?: string;
  email?: string;
  role?: string;
  password?: string;
}

export default function UserForm({
  mode,
  initialDisplayName = '',
  initialEmail = '',
  initialRole = 'user',
  initialUsername = '',
  isSubmitting,
  submitError,
  onSubmit,
  onCancel,
}: UserFormProps) {
  const { t } = useTranslation();
  const isCreate = mode === 'create';

  const [username, setUsername] = useState(initialUsername);
  const [displayName, setDisplayName] = useState(initialDisplayName);
  const [email, setEmail] = useState(initialEmail);
  const [role, setRole] = useState<UserRole>(initialRole);
  const [password, setPassword] = useState('');
  const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});

  const validate = (): boolean => {
    const errors: ValidationErrors = {};

    if (isCreate && !username.trim()) {
      errors.username = t('user.validation.usernameRequired');
    }
    if (!displayName.trim()) {
      errors.displayName = t('user.validation.displayNameRequired');
    }
    const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!email.trim()) {
      errors.email = t('user.validation.emailRequired');
    } else if (!EMAIL_RE.test(email.trim())) {
      errors.email = t('user.validation.emailInvalid');
    }
    if (!role) {
      errors.role = t('user.validation.roleRequired');
    }
    if (isCreate) {
      if (!password) {
        errors.password = t('user.validation.passwordRequired');
      } else if (password.length < 8) {
        errors.password = t('user.validation.passwordTooShort');
      }
    } else if (password && password.length < 8) {
      errors.password = t('user.validation.passwordTooShort');
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    if (isCreate) {
      const data: CreateUserRequest = {
        username: username.trim(),
        display_name: displayName.trim(),
        email: email.trim(),
        role,
        password,
      };
      onSubmit(data);
    } else {
      const data: UpdateUserRequest = {
        display_name: displayName.trim(),
        email: email.trim(),
        role,
      };
      if (password) {
        data.password = password;
      }
      onSubmit(data);
    }
  };

  return (
    <form onSubmit={handleSubmit} noValidate className="user-form">
      {isCreate && (
        <div className="form-field">
          <label htmlFor="user-username" className="form-label">
            {t('user.form.usernameLabel')}
          </label>
          <input
            id="user-username"
            type="text"
            className="form-input"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder={t('user.form.usernamePlaceholder')}
            aria-describedby={validationErrors.username ? 'user-username-error' : undefined}
            aria-invalid={!!validationErrors.username}
            disabled={isSubmitting}
          />
          {validationErrors.username && (
            <p id="user-username-error" role="alert" className="form-error">
              {validationErrors.username}
            </p>
          )}
        </div>
      )}

      {!isCreate && (
        <div className="form-field">
          <label htmlFor="user-username-readonly" className="form-label">
            {t('user.form.usernameLabel')}
          </label>
          <input
            id="user-username-readonly"
            type="text"
            className="form-input"
            value={initialUsername}
            readOnly
            aria-readonly="true"
          />
        </div>
      )}

      <div className="form-field">
        <label htmlFor="user-display-name" className="form-label">
          {t('user.form.displayNameLabel')}
        </label>
        <input
          id="user-display-name"
          type="text"
          className="form-input"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          placeholder={t('user.form.displayNamePlaceholder')}
          aria-describedby={validationErrors.displayName ? 'user-display-name-error' : undefined}
          aria-invalid={!!validationErrors.displayName}
          disabled={isSubmitting}
        />
        {validationErrors.displayName && (
          <p id="user-display-name-error" role="alert" className="form-error">
            {validationErrors.displayName}
          </p>
        )}
      </div>

      <div className="form-field">
        <label htmlFor="user-email" className="form-label">
          {t('user.form.emailLabel')}
        </label>
        <input
          id="user-email"
          type="email"
          className="form-input"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder={t('user.form.emailPlaceholder')}
          aria-describedby={validationErrors.email ? 'user-email-error' : undefined}
          aria-invalid={!!validationErrors.email}
          disabled={isSubmitting}
        />
        {validationErrors.email && (
          <p id="user-email-error" role="alert" className="form-error">
            {validationErrors.email}
          </p>
        )}
      </div>

      <div className="form-field">
        <label htmlFor="user-role" className="form-label">
          {t('user.form.roleLabel')}
        </label>
        <select
          id="user-role"
          className="form-input"
          value={role}
          onChange={(e) => setRole(e.target.value as UserRole)}
          aria-describedby={validationErrors.role ? 'user-role-error' : undefined}
          aria-invalid={!!validationErrors.role}
          disabled={isSubmitting}
        >
          {USER_ROLES.map((r) => (
            <option key={r} value={r}>
              {t(`user.role.${r}`)}
            </option>
          ))}
        </select>
        {validationErrors.role && (
          <p id="user-role-error" role="alert" className="form-error">
            {validationErrors.role}
          </p>
        )}
      </div>

      <div className="form-field">
        <label htmlFor="user-password" className="form-label">
          {t('user.form.passwordLabel')}
          {!isCreate && (
            <span className="form-hint"> — {t('user.form.passwordEditHint')}</span>
          )}
        </label>
        <input
          id="user-password"
          type="password"
          className="form-input"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder={t('user.form.passwordPlaceholder')}
          aria-describedby={validationErrors.password ? 'user-password-error' : undefined}
          aria-invalid={!!validationErrors.password}
          disabled={isSubmitting}
          autoComplete="new-password"
        />
        {validationErrors.password && (
          <p id="user-password-error" role="alert" className="form-error">
            {validationErrors.password}
          </p>
        )}
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
