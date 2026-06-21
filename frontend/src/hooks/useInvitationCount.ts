import { useState, useEffect, useCallback } from 'react';
import { fetchMyInvitations } from '../utils/invitationApi';

const POLL_INTERVAL_MS = 60_000; // 1分ごとに再取得

/**
 * 自分宛の pending 招待件数を返すカスタムフック。
 * NavBar でバッジ表示に使用する。
 */
export function useInvitationCount(): { count: number; refresh: () => void } {
  const [count, setCount] = useState(0);

  const fetchCount = useCallback(async () => {
    try {
      const all = await fetchMyInvitations();
      setCount(all.filter((inv) => inv.status === 'pending').length);
    } catch {
      // エラー時はカウントを維持する
    }
  }, []);

  useEffect(() => {
    void fetchCount();
    const timer = setInterval(() => {
      void fetchCount();
    }, POLL_INTERVAL_MS);
    return () => {
      clearInterval(timer);
    };
  }, [fetchCount]);

  return { count, refresh: fetchCount };
}
