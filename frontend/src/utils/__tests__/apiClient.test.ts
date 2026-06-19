import MockAdapter from 'axios-mock-adapter';
import apiClient from '../apiClient';
import * as authUtils from '../auth';

vi.mock('../auth', () => ({
  getToken: vi.fn(),
  clearAuthStorage: vi.fn(),
}));

const mockGetToken = vi.mocked(authUtils.getToken);
const mockClearAuthStorage = vi.mocked(authUtils.clearAuthStorage);

// jsdom では window.location.replace を spyOn で再定義できないため
// Object.defineProperty で差し替える
const locationReplaceMock = vi.fn();

Object.defineProperty(window, 'location', {
  value: {
    ...window.location,
    replace: locationReplaceMock,
  },
  writable: true,
});

let mockAxios: MockAdapter;

beforeEach(() => {
  mockAxios = new MockAdapter(apiClient);
  locationReplaceMock.mockReset();
});

afterEach(() => {
  mockAxios.restore();
  vi.restoreAllMocks();
  localStorage.clear();
});

describe('apiClient', () => {
  describe('リクエストインターセプター', () => {
    it('localStorageにトークンがある場合、Authorization: Bearer ヘッダーが付く', async () => {
      mockGetToken.mockReturnValue('test-token-123');
      mockAxios.onGet('/test').reply(200, { data: 'ok' });

      const response = await apiClient.get('/test');

      expect(response.config.headers['Authorization']).toBe('Bearer test-token-123');
    });

    it('トークンなしの場合、Authorization ヘッダーが付かない', async () => {
      mockGetToken.mockReturnValue(null);
      mockAxios.onGet('/test').reply(200, { data: 'ok' });

      const response = await apiClient.get('/test');

      expect(response.config.headers['Authorization']).toBeUndefined();
    });
  });

  describe('レスポンスインターセプター', () => {
    it('401 レスポンスで clearAuthStorage が呼ばれる', async () => {
      mockGetToken.mockReturnValue(null);
      mockAxios.onGet('/protected').reply(401, { error: 'Unauthorized' });

      await expect(apiClient.get('/protected')).rejects.toThrow();

      expect(mockClearAuthStorage).toHaveBeenCalled();
    });

    it('401 レスポンスで window.location.replace("/login") が呼ばれる', async () => {
      mockGetToken.mockReturnValue(null);
      mockAxios.onGet('/protected').reply(401, { error: 'Unauthorized' });

      await expect(apiClient.get('/protected')).rejects.toThrow();

      expect(locationReplaceMock).toHaveBeenCalledWith('/login');
    });

    it('200 レスポンスはそのまま返る', async () => {
      mockGetToken.mockReturnValue(null);
      mockAxios.onGet('/data').reply(200, { data: { id: '1', name: 'テスト' } });

      const response = await apiClient.get('/data');

      expect(response.status).toBe(200);
      expect(response.data).toEqual({ data: { id: '1', name: 'テスト' } });
    });
  });
});
