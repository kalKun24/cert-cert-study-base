/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      // バックエンドへのプロキシ設定
      // VITE_API_TARGET 環境変数で切り替え可能（Docker外では localhost:8080 を使用）
      '/api': {
        target: process.env.VITE_API_TARGET ?? 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
    // remark / rehype / unified 系は pure ESM のため vitest のトランスフォーム対象に含める
    server: {
      deps: {
        inline: [
          /react-markdown/,
          /remark-/,
          /rehype-/,
          /unified/,
          /bail/,
          /is-plain-obj/,
          /trough/,
          /vfile/,
          /unist-/,
          /hast-/,
          /mdast-/,
          /micromark/,
          /decode-named-character-reference/,
          /character-entities/,
          /property-information/,
          /space-separated-tokens/,
          /comma-separated-tokens/,
          /web-namespaces/,
          /zwitch/,
          /html-void-elements/,
          /stringify-entities/,
          /ccount/,
          /longest-streak/,
          /highlight\.js/,
          /katex/,
        ],
      },
    },
  },
});
