import { Tag } from '../types/tag';
import { useTranslation } from 'react-i18next';

interface TagChipProps {
  tag: Tag;
  selected: boolean;
  onToggle: (id: string) => void;
}

export default function TagChip({ tag, selected, onToggle }: TagChipProps) {
  const { t } = useTranslation();

  return (
    <button
      type="button"
      role="checkbox"
      aria-checked={selected}
      aria-label={selected ? `${tag.name} - ${t('common.selected')}` : tag.name}
      className={`tag-chip${selected ? ' tag-chip--selected' : ''}`}
      onClick={() => onToggle(tag.name)}
    >
      {tag.name}
    </button>
  );
}
