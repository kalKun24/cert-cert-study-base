import { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { TeamMemberStats } from '../types/team';
import { fetchTeamMemberStats } from '../utils/teamApi';

export default function TeamMemberListPage() {
  const { id } = useParams<{ id: string }>();
  const { t } = useTranslation();

  const [members, setMembers] = useState<TeamMemberStats[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState('');

  useEffect(() => {
    if (!id) return;
    setIsLoading(true);
    setFetchError('');
    fetchTeamMemberStats(id)
      .then(setMembers)
      .catch(() => setFetchError(t('team.members.error.fetchFailed')))
      .finally(() => setIsLoading(false));
  }, [id, t]);

  /**
   * last_login_at を日本語ロケールの日時形式に整形する。
   * null の場合はロケールキー team.members.noLastLogin の値を返す。
   */
  const formatLastLoginAt = (value: string | null): string => {
    if (value === null) return t('team.members.noLastLogin');
    return new Date(value).toLocaleString('ja-JP');
  };

  /**
   * role 文字列をロケールキーから翻訳する。
   * 定義外のロール値はそのまま表示する。
   */
  const formatRole = (role: string): string => {
    const key = `team.members.roles.${role}` as const;
    const translated = t(key);
    // i18next はキーが未定義でもキー文字列を返すため、変換結果をそのまま使用
    return translated;
  };

  return (
    <section className="page-container-full">
      <div className="page-header">
        <h1 className="page-title">{t('team.members.title')}</h1>
        <Link to={`/teams/${id}`} className="btn btn-secondary">
          ← {t('team.members.backToTeam')}
        </Link>
      </div>

      {isLoading ? (
        <p role="status" className="page-loading">
          {t('common.loading')}
        </p>
      ) : fetchError ? (
        <p role="alert" className="alert alert-error">
          {fetchError}
        </p>
      ) : members.length === 0 ? (
        <p>{t('team.members.empty')}</p>
      ) : (
        <div className="table-wrapper">
          <table className="user-table">
            <thead>
              <tr>
                <th scope="col">{t('team.members.columns.displayName')}</th>
                <th scope="col">{t('team.members.columns.role')}</th>
                <th scope="col">{t('team.members.columns.isTeamOwner')}</th>
                <th scope="col">{t('team.members.columns.questionCount')}</th>
                <th scope="col">{t('team.members.columns.commentCount')}</th>
                <th scope="col">{t('team.members.columns.lastLoginAt')}</th>
              </tr>
            </thead>
            <tbody>
              {members.map((member) => (
                <tr key={member.user_id}>
                  <td data-label={t('team.members.columns.displayName')}>
                    {member.display_name}
                  </td>
                  <td data-label={t('team.members.columns.role')}>
                    <span className="role-badge" data-role={member.role}>
                      {formatRole(member.role)}
                    </span>
                  </td>
                  <td data-label={t('team.members.columns.isTeamOwner')}>
                    {member.is_team_owner
                      ? t('team.members.isTeamOwner.yes')
                      : t('team.members.isTeamOwner.no')}
                  </td>
                  <td data-label={t('team.members.columns.questionCount')}>
                    {member.question_count}
                  </td>
                  <td data-label={t('team.members.columns.commentCount')}>
                    {member.comment_count}
                  </td>
                  <td data-label={t('team.members.columns.lastLoginAt')}>
                    {formatLastLoginAt(member.last_login_at)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
