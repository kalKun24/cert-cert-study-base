import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';

interface TeamOwnerRoleModalProps {
  targetUserName: string;
  action: 'grant' | 'revoke';
  onConfirm: () => void;
  onCancel: () => void;
}

export default function TeamOwnerRoleModal({
  targetUserName,
  action,
  onConfirm,
  onCancel,
}: TeamOwnerRoleModalProps) {
  const { t } = useTranslation();
  const cancelRef = useRef<HTMLButtonElement>(null);
  const titleId = 'team-owner-role-modal-title';

  // 初期フォーカスをキャンセルボタンに当てる
  useEffect(() => {
    cancelRef.current?.focus();
  }, []);

  // Escape キーでキャンセル
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onCancel();
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onCancel]);

  const isGrant = action === 'grant';
  const title = isGrant
    ? t('team.ownerModal.grantTitle')
    : t('team.ownerModal.revokeTitle');

  return (
    <>
      <div
        className="modal-overlay"
        aria-hidden="true"
        onClick={onCancel}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        className="modal-box"
      >
        <h2 id={titleId} className="modal-title">
          {title}
        </h2>

        <p className="modal-target-user">
          <span className="modal-field-label">{t('team.ownerModal.targetUser')}:</span>{' '}
          <strong>{targetUserName}</strong>
        </p>

        {isGrant ? (
          <>
            <div className="modal-capabilities-block">
              <p className="modal-field-label">{t('team.ownerModal.ownerCapabilities')}</p>
              <p className="modal-capabilities-text">
                {t('team.ownerModal.capabilitiesList')}
              </p>
            </div>

            <p role="alert" className="alert alert-error modal-abuse-warning">
              {t('team.ownerModal.abuseWarning')}
            </p>
          </>
        ) : (
          <p className="modal-revoke-confirm">
            {t('team.ownerModal.revokeConfirm', { name: targetUserName })}
          </p>
        )}

        <div className="modal-actions">
          <button
            ref={cancelRef}
            type="button"
            className="btn btn-secondary"
            onClick={onCancel}
          >
            {t('team.ownerModal.cancelButton')}
          </button>
          <button
            type="button"
            className={isGrant ? 'btn btn-primary' : 'btn btn-danger'}
            onClick={onConfirm}
          >
            {isGrant ? t('team.ownerModal.grantButton') : t('team.ownerModal.revokeButton')}
          </button>
        </div>
      </div>
    </>
  );
}
