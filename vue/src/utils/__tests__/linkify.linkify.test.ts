import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import titleExist from '@/utils/mediawiki';
import linkify from '../linkify';

vi.mock('@/utils/mediawiki', () => ({
  default: vi.fn(),
}));

describe('linkify', () => {
  let errorSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    errorSpy = vi.spyOn(console, 'error').mockImplementation(() => { });
  });

  afterEach(() => {
    errorSpy.mockRestore();
    vi.clearAllMocks();
  });


  it('should convert URLs into anchor tags', async () => {
    expect(await linkify('Visit http://example.com for more info'))
      .toBe('Visit <a href="http://example.com" class="external external-url" target="_blank" rel="noopener noreferrer">http://example.com</a> for more info');
  });

  it('should not modify text without URLs or wiki links', async () => {
    expect(await linkify('This is a simple text without links.'))
      .toBe('This is a simple text without links.');
  });

  it('should handle multiple URLs and wiki links in the same string', async () => {
    vi.mocked(titleExist).mockResolvedValue(true);
    expect(await linkify('Check http://example.com and [[TestPage]]'))
      .toBe('Check <a href="http://example.com" class="external external-url" target="_blank" rel="noopener noreferrer">http://example.com</a> and <a href="/wiki/TestPage" class="internal">TestPage</a>');
  });

  it('should convert non-existing wiki links with new class', async () => {
    vi.mocked(titleExist).mockResolvedValue(false);
    expect(await linkify('Check [[MissingPage]]'))
      .toBe('Check <a href="/wiki/MissingPage" class="internal new">MissingPage</a>');
  });

  it('should sanitize input before processing', async () => {
    expect(await linkify('<script>alert("XSS")</script>Visit http://example.com'))
      .not.toContain('<script>');
  });
});
