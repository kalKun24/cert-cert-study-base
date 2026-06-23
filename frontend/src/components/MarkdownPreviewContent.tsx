import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import remarkMath from 'remark-math';
import rehypeHighlight from 'rehype-highlight';
import rehypeKatex from 'rehype-katex';
import rehypeSanitize, { defaultSchema } from 'rehype-sanitize';

// KaTeX が生成する数学タグ・クラスを許可しつつ sanitize を有効にするスキーマ。
// defaultSchema をベースに KaTeX 出力に必要なタグと属性を追加する。
const katexTagNames = [
  'math', 'mi', 'mn', 'mo', 'ms', 'mspace', 'mtext',
  'mfrac', 'mroot', 'msqrt', 'mtable', 'mtr', 'mtd',
  'mover', 'munder', 'munderover', 'msup', 'msub', 'msubsup',
  'semantics', 'annotation',
];

const sanitizeSchema = {
  ...defaultSchema,
  attributes: {
    ...defaultSchema.attributes,
    '*': [
      ...((defaultSchema.attributes?.['*'] as string[] | undefined) ?? []),
      'className',
      'style',
    ],
    span: [
      ...((defaultSchema.attributes?.['span'] as string[] | undefined) ?? []),
      'aria-hidden',
    ],
  },
  tagNames: [
    ...((defaultSchema.tagNames as string[] | undefined) ?? []),
    ...katexTagNames,
  ],
};

export interface MarkdownPreviewContentProps {
  value: string;
  emptyMessage?: string;
  className?: string;
}

export default function MarkdownPreviewContent({
  value,
  emptyMessage,
  className = 'markdown-content',
}: MarkdownPreviewContentProps) {
  if (!value.trim()) {
    return <p className="md-editor-preview-empty">{emptyMessage}</p>;
  }

  return (
    <div className={className}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkBreaks, remarkMath]}
        rehypePlugins={[rehypeHighlight, rehypeKatex, [rehypeSanitize, sanitizeSchema]]}
      >
        {value}
      </ReactMarkdown>
    </div>
  );
}
