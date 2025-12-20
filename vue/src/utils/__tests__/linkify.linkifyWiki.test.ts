import { afterEach,beforeEach, describe, expect, it, vi } from 'vitest';

import { linkifyWiki } from '@/utils/linkify';
import titleExist from '@/utils/mediawiki';

vi.mock('@/utils/mediawiki', () => ({
  default: vi.fn(),
}));

describe('linkifyWiki', () => {
  let errorSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    errorSpy = vi.spyOn(console, 'error').mockImplementation(() => { });
  });

  afterEach(() => {
    errorSpy.mockRestore();
    vi.clearAllMocks();
  });

  it('should convert an existing wiki link', async () => {
    vi.mocked(titleExist).mockResolvedValue(true);
    const result = await linkifyWiki('This is a [[TestPage]] link');
    expect(result).toBe('This is a <a href="/wiki/TestPage" class="internal">TestPage</a> link');
  });

  it('should convert a non-existing wiki link', async () => {
    vi.mocked(titleExist).mockResolvedValue(false);
    const result = await linkifyWiki('This is a [[MissingPage]] link');
    expect(result).toBe('This is a <a href="/wiki/MissingPage" class="internal new">MissingPage</a> link');
  });

  it('should handle multiple same links', async () => {
    vi.mocked(titleExist).mockResolvedValue(true);
    const result = await linkifyWiki('[[TestPage]] is linked twice: [[TestPage]].');
    expect(result).toBe('<a href="/wiki/TestPage" class="internal">TestPage</a> is linked twice: <a href="/wiki/TestPage" class="internal">TestPage</a>.');
  });

  it('should handle multiple different links', async () => {
    vi.mocked(titleExist).mockImplementation(async () => Boolean([true, false].shift()));
    const result = await linkifyWiki('Links: [[TestPage1]], [[TestPage2]].');
    expect(result).toBe('Links: <a href="/wiki/TestPage1" class="internal">TestPage1</a>, <a href="/wiki/TestPage2" class="internal">TestPage2</a>.');
  });

  it('should fallback on API failure', async () => {
    vi.mocked(titleExist).mockResolvedValue(false);
    const result = await linkifyWiki('This is a [[ErrorPage]] link');
    expect(result).toBe('This is a <a href="/wiki/ErrorPage" class="internal new">ErrorPage</a> link');
  });
});
