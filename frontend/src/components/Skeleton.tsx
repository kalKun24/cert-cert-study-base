/** スケルトンローディングコンポーネント */

interface SkeletonProps {
  /** 幅（CSS値を文字列で指定。例: "100%", "200px"） */
  width?: string;
  /** 高さ（CSS値を文字列で指定。例: "1rem", "40px"） */
  height?: string;
  /** 形状バリアント */
  variant?: 'text' | 'rect' | 'circle';
  /** 追加のクラス名 */
  className?: string;
}

/**
 * shimmerアニメーション付きのスケルトンローディングコンポーネント。
 * ローディング中にコンテンツのプレースホルダーとして使用する。
 */
export default function Skeleton({
  width,
  height,
  variant = 'rect',
  className = '',
}: SkeletonProps) {
  const style: React.CSSProperties = {};
  if (width) style.width = width;
  if (height) style.height = height;

  return (
    <span
      className={`skeleton skeleton--${variant} ${className}`.trim()}
      style={style}
      aria-hidden="true"
    />
  );
}

/** 問題一覧のスケルトンアイテム */
export function QuestionListSkeleton() {
  return (
    <ul className="question-list" aria-busy="true" aria-label="読み込み中">
      {Array.from({ length: 5 }).map((_, i) => (
        <li key={i} className="question-list-item">
          <div className="skeleton-question-item">
            <Skeleton variant="text" width="70%" height="1.1rem" />
            <div className="skeleton-question-meta">
              <Skeleton variant="text" width="5rem" height="0.75rem" />
              <Skeleton variant="rect" width="4rem" height="1.25rem" />
              <Skeleton variant="rect" width="4rem" height="1.25rem" />
            </div>
          </div>
        </li>
      ))}
    </ul>
  );
}

/** 問題詳細のスケルトン */
export function QuestionDetailSkeleton() {
  return (
    <div className="question-detail" aria-busy="true" aria-label="読み込み中">
      <div className="skeleton-detail-header">
        <Skeleton variant="text" width="4rem" height="0.875rem" />
        <Skeleton variant="text" width="60%" height="1.875rem" />
        <div className="skeleton-question-meta">
          <Skeleton variant="text" width="5rem" height="0.75rem" />
          <Skeleton variant="rect" width="3.5rem" height="1.25rem" />
        </div>
      </div>
      <div className="skeleton-section">
        <Skeleton variant="text" width="5rem" height="1.125rem" />
        <Skeleton variant="rect" width="100%" height="6rem" />
      </div>
      <div className="skeleton-section">
        <Skeleton variant="text" width="3rem" height="1.125rem" />
        <Skeleton variant="rect" width="100%" height="4rem" />
      </div>
    </div>
  );
}

/** ホームページのダッシュボードスケルトン */
export function DashboardSkeleton() {
  return (
    <div aria-busy="true" aria-label="読み込み中">
      <div className="skeleton-stat-cards">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="skeleton-stat-card">
            <Skeleton variant="text" width="60%" height="0.875rem" />
            <Skeleton variant="text" width="40%" height="2rem" />
          </div>
        ))}
      </div>
      <div className="skeleton-section">
        <Skeleton variant="text" width="8rem" height="1.25rem" />
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} variant="rect" width="100%" height="3.5rem" />
        ))}
      </div>
    </div>
  );
}
