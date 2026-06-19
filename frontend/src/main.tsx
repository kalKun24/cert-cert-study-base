import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import App from './App';
import ja from './locales/ja.json';
import './styles/global.css';

// i18next の初期化
i18n.use(initReactI18next).init({
  resources: {
    ja: {
      translation: ja,
    },
  },
  lng: 'ja',
  fallbackLng: 'ja',
  interpolation: {
    // React は XSS をデフォルトでエスケープするため無効化
    escapeValue: false,
  },
});

const rootElement = document.getElementById('root');
if (!rootElement) {
  throw new Error('ルート要素が見つかりません。index.html に id="root" の要素が必要です。');
}

createRoot(rootElement).render(
  <StrictMode>
    <App />
  </StrictMode>
);
