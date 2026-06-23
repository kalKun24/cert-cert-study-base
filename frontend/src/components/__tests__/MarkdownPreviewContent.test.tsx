import { render } from '@testing-library/react';
import MarkdownPreviewContent from '../MarkdownPreviewContent';

describe('MarkdownPreviewContent', () => {
  describe('基本レンダリング', () => {
    it('value が空の場合 emptyMessage が表示されること', () => {
      const { getByText } = render(
        <MarkdownPreviewContent value="" emptyMessage="テキストがありません" />,
      );
      expect(getByText('テキストがありません')).toBeInTheDocument();
    });

    it('value がある場合 emptyMessage が表示されないこと', () => {
      const { queryByText } = render(
        <MarkdownPreviewContent value="# 見出し" emptyMessage="テキストがありません" />,
      );
      expect(queryByText('テキストがありません')).not.toBeInTheDocument();
    });

    it('引用ブロックが blockquote 要素としてレンダリングされること', () => {
      const { container } = render(<MarkdownPreviewContent value="> 引用テキスト" />);
      expect(container.querySelector('blockquote')).toBeInTheDocument();
    });

    it('斜体テキストが em 要素としてレンダリングされること', () => {
      const { container } = render(<MarkdownPreviewContent value="*斜体テキスト*" />);
      const em = container.querySelector('em');
      expect(em).toBeInTheDocument();
      expect(em?.textContent).toBe('斜体テキスト');
    });

    it('Enter 1 回の改行がプレビューに br 要素として反映されること（remark-breaks）', () => {
      const { container } = render(<MarkdownPreviewContent value={'1行目\n2行目'} />);
      expect(container.querySelector('br')).toBeInTheDocument();
    });
  });

  describe('GFM（GitHub Flavored Markdown）', () => {
    it('テーブルが table > thead > th / tbody > td としてレンダリングされること', () => {
      const md = '| 列1 | 列2 |\n| --- | --- |\n| セルA | セルB |';
      const { container } = render(<MarkdownPreviewContent value={md} />);
      expect(container.querySelector('table')).toBeInTheDocument();
      expect(container.querySelector('th')).toBeInTheDocument();
      expect(container.querySelector('td')).toBeInTheDocument();
    });

    it('取り消し線が del 要素としてレンダリングされること', () => {
      const { container } = render(<MarkdownPreviewContent value="~~取り消し~~" />);
      const del = container.querySelector('del');
      expect(del).toBeInTheDocument();
      expect(del?.textContent).toBe('取り消し');
    });

    it('未完了チェックリストが unchecked checkbox としてレンダリングされること', () => {
      const { container } = render(<MarkdownPreviewContent value="- [ ] 未完了" />);
      const checkbox = container.querySelector('input[type="checkbox"]');
      expect(checkbox).toBeInTheDocument();
      expect(checkbox).not.toBeChecked();
    });

    it('完了チェックリストが checked checkbox としてレンダリングされること', () => {
      const { container } = render(<MarkdownPreviewContent value="- [x] 完了" />);
      const checkbox = container.querySelector('input[type="checkbox"]');
      expect(checkbox).toBeInTheDocument();
      expect(checkbox).toBeChecked();
    });

    it('順序なしリストが ul > li としてレンダリングされること', () => {
      const md = '- 項目A\n- 項目B\n- 項目C';
      const { container } = render(<MarkdownPreviewContent value={md} />);
      expect(container.querySelector('ul')).toBeInTheDocument();
      expect(container.querySelectorAll('li')).toHaveLength(3);
    });

    it('順序ありリストが ol > li としてレンダリングされること', () => {
      const md = '1. 手順1\n2. 手順2';
      const { container } = render(<MarkdownPreviewContent value={md} />);
      expect(container.querySelector('ol')).toBeInTheDocument();
      expect(container.querySelectorAll('li')).toHaveLength(2);
    });
  });

  describe('シンタックスハイライト', () => {
    it('言語指定ありコードブロックに language-* クラスが付与されること', () => {
      const md = '```python\nprint("hello")\n```';
      const { container } = render(<MarkdownPreviewContent value={md} />);
      const code = container.querySelector('pre code');
      expect(code).toBeInTheDocument();
      expect(code?.className).toMatch(/language-python/);
    });

    it('言語指定なしコードブロックも pre > code としてレンダリングされること', () => {
      const md = '```\nsome code\n```';
      const { container } = render(<MarkdownPreviewContent value={md} />);
      expect(container.querySelector('pre code')).toBeInTheDocument();
    });
  });

  describe('数式（KaTeX）', () => {
    it('インライン数式が KaTeX によってレンダリングされること', () => {
      const { container } = render(<MarkdownPreviewContent value="$E = mc^2$" />);
      // rehype-katex は .katex クラスを付与した span を生成する
      expect(container.querySelector('.katex')).toBeInTheDocument();
    });

    it('ブロック数式が KaTeX によってレンダリングされること', () => {
      const md = '$$\n\\sum_{i=1}^{n} i\n$$';
      const { container } = render(<MarkdownPreviewContent value={md} />);
      expect(container.querySelector('.katex-display')).toBeInTheDocument();
    });
  });

  describe('セキュリティ（rehype-sanitize）', () => {
    it('javascript: URL がリンクの href から除去されること', () => {
      const { container } = render(
        <MarkdownPreviewContent value="[click](javascript:alert(1))" />,
      );
      const link = container.querySelector('a');
      // rehype-sanitize は javascript: URL を href ごと除去する（null になる）か、プロトコルを取り除く
      const href = link?.getAttribute('href') ?? '';
      expect(href).not.toMatch(/^javascript:/i);
    });

    it('script タグが出力に含まれないこと', () => {
      const { container } = render(
        <MarkdownPreviewContent value={'<script>alert("xss")</script>'} />,
      );
      expect(container.querySelector('script')).not.toBeInTheDocument();
    });

    it('onerror などのイベントハンドラ属性が img タグから除去されること', () => {
      const { container } = render(
        <MarkdownPreviewContent value={'<img src="x" onerror="alert(1)">'} />,
      );
      const img = container.querySelector('img');
      if (img) {
        expect(img.getAttribute('onerror')).toBeNull();
      }
    });

    it('数式（KaTeX）が sanitize 後も正しくレンダリングされること（スキーマ確認）', () => {
      const { container } = render(<MarkdownPreviewContent value="$E = mc^2$" />);
      // KaTeX の className が sanitize で除去されていないこと
      expect(container.querySelector('.katex')).toBeInTheDocument();
    });

    it('ブロック数式が sanitize 後も正しくレンダリングされること', () => {
      const md = '$$\n\\frac{a}{b}\n$$';
      const { container } = render(<MarkdownPreviewContent value={md} />);
      expect(container.querySelector('.katex-display')).toBeInTheDocument();
    });

    it('className prop がラッパー div に適用されること', () => {
      const { container } = render(
        <MarkdownPreviewContent value="テスト" className="custom-class" />,
      );
      expect(container.querySelector('.custom-class')).toBeInTheDocument();
    });

    it('className 省略時はデフォルト markdown-content クラスが適用されること', () => {
      const { container } = render(<MarkdownPreviewContent value="テスト" />);
      expect(container.querySelector('.markdown-content')).toBeInTheDocument();
    });
  });
});
