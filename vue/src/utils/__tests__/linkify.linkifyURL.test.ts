import { describe, expect,it } from 'vitest';

import linkifyURL from '../linkify';

describe('linkifyURL', () => {
  it('should convert URLs into anchor tags', async () => {
    expect(await linkifyURL('Visit http://example.com for more info'))
      .toBe('Visit <a href="http://example.com" class="external external-url" target="_blank" rel="noopener noreferrer">http://example.com</a> for more info');
  });

  it('should not modify text without URLs', async () => {
    expect(await linkifyURL('This is a simple text without links.'))
      .toBe('This is a simple text without links.');
  });

  it('should handle multiple URLs in the same string', async () => {
    expect(await linkifyURL('Check http://example.com and https://test.com'))
      .toBe('Check <a href="http://example.com" class="external external-url" target="_blank" rel="noopener noreferrer">http://example.com</a> and <a href="https://test.com" class="external external-url" target="_blank" rel="noopener noreferrer">https://test.com</a>');
  });

  it('should preserve surrounding text', async () => {
    expect(await linkifyURL('Go to http://example.com now!'))
      .toBe('Go to <a href="http://example.com" class="external external-url" target="_blank" rel="noopener noreferrer">http://example.com</a> now!');
  });
});
