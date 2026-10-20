import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';

describe('API documentation', () => {
  it('uses the same specification as the panel', () => {
    const panel = readFileSync(
      new URL('../../frontend/public/openapi.json', import.meta.url),
      'utf8',
    );
    const docs = readFileSync(new URL('../public/openapi.json', import.meta.url), 'utf8');
    expect(docs).toBe(panel);
  });
});
