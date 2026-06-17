import { useTranslation } from 'react-i18next';

/**
 * アプリケーションのルートコンポーネント。
 * 現段階ではウェルカム画面を表示する。
 * 今後、ルーティングやレイアウトをここに追加する。
 */
function App() {
  const { t } = useTranslation();

  return (
    <main>
      <h1>{t('app.title')}</h1>
      <p>{t('app.description')}</p>
      <p>{t('app.tagline')}</p>
    </main>
  );
}

export default App;
