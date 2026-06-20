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
  const itemRefs = useRef<(HTMLLIElement | null)[]>([]);

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

  /**
   * リスト項目内での矢印キーナビゲーション（WCAG 2.1 SC 2.4.7 対応）
   * ArrowDown: 次の項目、ArrowUp: 前の項目、Home: 先頭、End: 末尾
   */
  const handleItemKeyDown = (e: React.KeyboardEvent<HTMLLIElement>, index: number) => {
    const items = itemRefs.current.filter((el): el is HTMLLIElement => el !== null);
    const last = items.length - 1;

    switch (e.key) {
      case 'Enter':
      case ' ':
        e.preventDefault();
        onToggle(tags[index].name);
        break;
      case 'ArrowDown':
        e.preventDefault();
        items[index < last ? index + 1 : 0]?.focus();
        break;
      case 'ArrowUp':
        e.preventDefault();
        items[index > 0 ? index - 1 : last]?.focus();
        break;
      case 'Home':
        e.preventDefault();
        items[0]?.focus();
        break;
      case 'End':
        e.preventDefault();
        items[last]?.focus();
        break;
      default:
        break;
    }
  };

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
          {tags.map((tag, index) => {
            const isSelected = selectedTagNames.includes(tag.name);
            return (
              <li
                key={tag.id}
                ref={(el) => { itemRefs.current[index] = el; }}
                role="option"
                aria-selected={isSelected}
                className={`tag-dropdown-item${isSelected ? ' tag-dropdown-item--selected' : ''}`}
                onClick={() => onToggle(tag.name)}
                onKeyDown={(e) => handleItemKeyDown(e, index)}
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
