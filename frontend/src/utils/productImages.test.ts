import { describe, it, expect } from 'vitest';
import { getProductImageUrl, getCategoryEmoji } from './productImages';

describe('getProductImageUrl', () => {
  it('returns a URL for known product IDs', () => {
    const url = getProductImageUrl('1');
    expect(url).toBeDefined();
    expect(url).toContain('prod-mercadona.imgix.net');
  });

  it('returns undefined for unknown product IDs', () => {
    expect(getProductImageUrl('9999')).toBeUndefined();
  });

  it('returns distinct URLs for different products', () => {
    const url1 = getProductImageUrl('1');
    const url2 = getProductImageUrl('2');
    expect(url1).not.toBe(url2);
  });
});

describe('getCategoryEmoji', () => {
  it('returns an emoji for known categories', () => {
    expect(getCategoryEmoji('Lácteos')).toBe('🥛');
    expect(getCategoryEmoji('Frutas')).toBe('🍎');
    expect(getCategoryEmoji('Bebidas')).toBe('🍺');
  });

  it('returns the fallback emoji for unknown categories', () => {
    expect(getCategoryEmoji('Desconocida')).toBe('🛒');
  });

  it('returns the fallback emoji when category is undefined', () => {
    expect(getCategoryEmoji(undefined)).toBe('🛒');
  });
});
