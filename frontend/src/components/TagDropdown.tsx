import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Tag } from '../types/tag';

interface TagDropdownProps {
  /** 選択可能なタグの一覧 */
  tags: Tag[];
  /** 選択中のタグ名の配列 */
  selectedTagNames: string[];
  /** タグのトグル時に呼ばれるコールバック */
  onToggle: (tagName: string) => void;
}

/**
 * タグ選択ドロップダウンコンポーネント
 * メタバー内でタグを複数選択するためのドロップダウン
 */
export default function TagDropdown({ tags, selectedTagNames, onToggle }: TagDropdownProps) {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // ドロップダウン外クリックで閉じる
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  // ESCキーでドロップダウンを閉じる
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener('keydown', handleKeyDown);
    }
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen]);

  // ボタンに表示するラベル（選択中タグをカンマ区切り or デフォルトラベル）
  const buttonLabel =
    selectedTagNames.length > 0
      ? selectedTagNames.join(', ')
      : t('question.tagSelectLabel');

  return (
    <div className="tag-dropdown" ref={containerRef}>
      <button
        type="button"
        className="tag-dropdown-trigger"
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        onClick={() => setIsOpen((prev) => !prev)}
      >
        <span className="tag-dropdown-label">{buttonLabel}</span>
        <span className="tag-dropdown-arrow" aria-hidden="true">▼</span>
      </button>

      {isOpen && (
        <ul
          className="tag-dropdown-menu"
          role="listbox"
          aria-multiselectable="true"
          aria-label={t('question.tagSelectLabel')}
        >
          {tags.length === 0 && (
            <li className="tag-dropdown-empty" role="option" aria-selected={false}>
              {t('tag.list.empty')}
            </li>
          )}
          {tags.map((tag) => {
            const isSelected = selectedTagNames.includes(tag.name);
            return (
              <li
                key={tag.id}
                role="option"
                aria-selected={isSelected}
                className={`tag-dropdown-item${isSelected ? ' tag-dropdown-item--selected' : ''}`}
                onClick={() => onToggle(tag.name)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    onToggle(tag.name);
                  }
                }}
                tabIndex={0}
              >
                <span className="tag-dropdown-checkbox" aria-hidden="true">
                  {isSelected ? '☑' : '☐'}
                </span>
                {tag.name}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
