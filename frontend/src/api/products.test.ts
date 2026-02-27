import { describe, it, expect, vi, beforeEach } from 'vitest';
import { searchProducts, getProduct } from './products';
import type { SearchResult, Product } from '../types';

const mockSearchResults: SearchResult[] = [
  { id: '1', name: 'LECHE ENTERA HACENDADO 1L', category: 'Lácteos', currentPrice: 0.89, minPrice: 0.79, maxPrice: 0.89 },
];

const mockProduct: Product = {
  id: '1',
  name: 'LECHE ENTERA HACENDADO 1L',
  category: 'Lácteos',
  currentPrice: 0.89,
  priceHistory: [
    { date: '2025-01-15T00:00:00Z', price: 0.79, store: 'Mercadona' },
    { date: '2025-09-22T00:00:00Z', price: 0.89, store: 'Mercadona' },
  ],
};

beforeEach(() => {
  vi.restoreAllMocks();
});

describe('searchProducts', () => {
  it('returns results when the response is OK', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockSearchResults),
    }));

    const results = await searchProducts('leche');
    expect(results).toEqual(mockSearchResults);
  });

  it('calls the correct endpoint with the encoded query', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([]),
    });
    vi.stubGlobal('fetch', fetchMock);

    await searchProducts('aceite oliva');
    expect(fetchMock).toHaveBeenCalledWith('/api/products?q=aceite%20oliva');
  });

  it('throws an Error when the response is not OK', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      statusText: 'Internal Server Error',
    }));

    await expect(searchProducts('leche')).rejects.toThrow('Search failed: Internal Server Error');
  });

  it('returns an empty array when the server returns []', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([]),
    }));

    const results = await searchProducts('xyznonexistent');
    expect(results).toEqual([]);
  });
});

describe('getProduct', () => {
  it('returns the product when the response is OK', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockProduct),
    }));

    const product = await getProduct('1');
    expect(product).toEqual(mockProduct);
  });

  it('calls the correct endpoint with the ID', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockProduct),
    });
    vi.stubGlobal('fetch', fetchMock);

    await getProduct('42');
    expect(fetchMock).toHaveBeenCalledWith('/api/products/42');
  });

  it('throws an Error when the product is not found', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      statusText: 'Not Found',
    }));

    await expect(getProduct('9999')).rejects.toThrow('Product not found: Not Found');
  });
});
