import { useTranslation } from 'react-i18next';

interface PaginatorProps {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export default function Paginator({ page, totalPages, onPageChange }: PaginatorProps) {
  const { t } = useTranslation();

  if (totalPages <= 1) return null;

  return (
    <nav
      className="pagination"
      role="navigation"
      aria-label={t('question.pagination.navLabel')}
    >
      <button
        type="button"
        className="btn btn-secondary"
        onClick={() => onPageChange(page - 1)}
        disabled={page <= 1}
        aria-label={t('question.pagination.prev')}
      >
        {t('question.pagination.prev')}
      </button>
      <span className="pagination-info">
        {t('question.pagination.pageInfo', { page, total: totalPages })}
      </span>
      <button
        type="button"
        className="btn btn-secondary"
        onClick={() => onPageChange(page + 1)}
        disabled={page >= totalPages}
        aria-label={t('question.pagination.next')}
      >
        {t('question.pagination.next')}
      </button>
    </nav>
  );
}
