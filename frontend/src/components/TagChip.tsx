import { Tag } from '../types/tag';
import { useTranslation } from 'react-i18next';

interface TagChipProps {
  tag: Tag;
  selected: boolean;
  onToggle: (id: string) => void;
}

export default function TagChip({ tag, selected, onToggle }: TagChipProps) {
  const { t } = useTranslation();

  const handleKeyDown = (e: React.KeyboardEvent<HTMLButtonElement>) => {
    // Enter・Spaceキーのデフォルト動作はbutton要素が処理するが
    // role="checkbox" との組み合わせで明示的に処理
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onToggle(tag.name);
    }
  };

  return (
    <button
      type="button"
      role="checkbox"
      aria-checked={selected}
      aria-label={selected ? `${tag.name} - ${t('common.selected')}` : tag.name}
      className={`tag-chip${selected ? ' tag-chip--selected' : ''}`}
      onClick={() => onToggle(tag.name)}
      onKeyDown={handleKeyDown}
    >
      {tag.name}
    </button>
  );
}
