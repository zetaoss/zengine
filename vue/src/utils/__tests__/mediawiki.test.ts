import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import http from '@/utils/http';
import titleExist from '@/utils/mediawiki';

describe('titleExist', () => {
  let errorSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    errorSpy = vi.spyOn(console, 'error').mockImplementation(() => { });
  });

  afterEach(() => {
    errorSpy.mockRestore();
  });

  it('returns true for existing title', async () => {
    vi.spyOn(http, 'get').mockResolvedValue({
      data: {
        query: {
          pages: {
            '12345': { pageid: 12345, title: 'Existing Title' },
          },
        },
      },
    });

    const result = await titleExist('Existing Title');
    expect(result).toBe(true);
  });

  it('returns false for non-existing title', async () => {
    vi.spyOn(http, 'get').mockResolvedValue({
      data: {
        query: {
          pages: {
            '-1': {},
          },
        },
      },
    });

    const result = await titleExist('Nonexistent Title');
    expect(result).toBe(false);
  });

  it('returns false on API failure', async () => {
    vi.spyOn(http, 'get').mockRejectedValue(new Error('Network error'));

    const result = await titleExist('Any Title');
    expect(result).toBe(false);
  });
});
