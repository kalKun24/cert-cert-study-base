import { useState, useId, useRef } from 'react';

interface AccordionSectionProps {
  title: string;
  children: React.ReactNode;
  /** デフォルトで開いた状態にするか（デフォルト: false = 閉じた状態） */
  defaultOpen?: boolean;
  /** セクション見出しのレベル（デフォルト: h2） */
  headingLevel?: 'h2' | 'h3' | 'h4';
  /** 追加のクラス名 */
  className?: string;
}

/**
 * アコーディオン形式の折りたたみセクションコンポーネント。
 * キーボード操作（Enter・Space）対応、aria-expanded / aria-controls 付き。
 */
export default function AccordionSection({
  title,
  children,
  defaultOpen = false,
  headingLevel: HeadingTag = 'h2',
  className = '',
}: AccordionSectionProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  const panelId = useId();
  const buttonId = useId();
  const contentRef = useRef<HTMLDivElement>(null);

  const toggle = () => setIsOpen((prev) => !prev);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLButtonElement>) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      toggle();
    }
  };

  return (
    <div className={`accordion-section ${className}`.trim()}>
      <HeadingTag className="accordion-heading">
        <button
          id={buttonId}
          type="button"
          className="accordion-trigger"
          aria-expanded={isOpen}
          aria-controls={panelId}
          onClick={toggle}
          onKeyDown={handleKeyDown}
        >
          <span className="accordion-title">{title}</span>
          <span className="accordion-icon" aria-hidden="true">
            {/* 三角形SVGアイコン: 開閉状態でCSSで回転させる */}
            <svg
              width="12"
              height="12"
              viewBox="0 0 12 12"
              fill="currentColor"
              className={`accordion-chevron${isOpen ? ' accordion-chevron--open' : ''}`}
            >
              <path d="M6 8L1 3h10L6 8z" />
            </svg>
          </span>
        </button>
      </HeadingTag>

      <div
        id={panelId}
        role="region"
        aria-labelledby={buttonId}
        className={`accordion-panel${isOpen ? ' accordion-panel--open' : ''}`}
        ref={contentRef}
      >
        <div className="accordion-panel-inner">{children}</div>
      </div>
    </div>
  );
}
