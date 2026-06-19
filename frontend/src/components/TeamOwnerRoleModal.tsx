import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';

interface TeamOwnerRoleModalProps {
  targetUserName: string;
  action: 'grant' | 'revoke';
  onConfirm: () => void;
  onCancel: () => void;
  isLoading?: boolean;
}

export default function TeamOwnerRoleModal({
  targetUserName,
  action,
  onConfirm,
  onCancel,
  isLoading = false,
}: TeamOwnerRoleModalProps) {
  const { t } = useTranslation();
  const cancelRef = useRef<HTMLButtonElement>(null);
  const modalRef = useRef<HTMLDivElement>(null);
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

  // フォーカストラップ
  useEffect(() => {
    const modalEl = modalRef.current;
    if (!modalEl) return;

    const focusableSelectors =
      'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

    const handleTabKey = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;
      const focusableElements = Array.from(
        modalEl.querySelectorAll<HTMLElement>(focusableSelectors),
      ).filter((el) => el.offsetParent !== null);

      if (focusableElements.length === 0) return;

      const firstEl = focusableElements[0];
      const lastEl = focusableElements[focusableElements.length - 1];

      if (e.shiftKey) {
        if (document.activeElement === firstEl) {
          e.preventDefault();
          lastEl.focus();
        }
      } else {
        if (document.activeElement === lastEl) {
          e.preventDefault();
          firstEl.focus();
        }
      }
    };

    document.addEventListener('keydown', handleTabKey);
    return () => document.removeEventListener('keydown', handleTabKey);
  }, []);

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
        ref={modalRef}
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
            disabled={isLoading}
            aria-busy={isLoading}
          >
            {isGrant ? t('team.ownerModal.grantButton') : t('team.ownerModal.revokeButton')}
          </button>
        </div>
      </div>
    </>
  );
}
