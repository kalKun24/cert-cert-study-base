import { Tag } from '../types/tag';

interface TagChipProps {
  tag: Tag;
  selected: boolean;
  onToggle: (id: string) => void;
}

export default function TagChip({ tag, selected, onToggle }: TagChipProps) {
  return (
    <button
      type="button"
      role="checkbox"
      aria-checked={selected}
      className={`tag-chip${selected ? ' tag-chip--selected' : ''}`}
      onClick={() => onToggle(tag.name)}
    >
      {tag.name}
    </button>
  );
}
