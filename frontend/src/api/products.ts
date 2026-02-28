import type { SearchResult, Product, TicketUploadResult, TicketUploadSummary, AnalyticsResult } from '../types';

const API_BASE = '/api';

export async function searchProducts(query: string): Promise<SearchResult[]> {
  const res = await fetch(`${API_BASE}/products?q=${encodeURIComponent(query)}`);
  if (!res.ok) {
    throw new Error(`Search failed: ${res.statusText}`);
  }
  return res.json();
}

export async function getAllProducts(): Promise<SearchResult[]> {
  return searchProducts('');
}

export async function getProduct(id: string): Promise<Product> {
  const res = await fetch(`${API_BASE}/products/${id}`);
  if (!res.ok) {
    throw new Error(`Product not found: ${res.statusText}`);
  }
  return res.json();
}

export async function uploadTicket(file: File): Promise<TicketUploadResult> {
  const form = new FormData();
  form.append('file', file);
  const res = await fetch(`${API_BASE}/tickets`, { method: 'POST', body: form });
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(text || res.statusText);
  }
  return res.json();
}

/**
 * Uploads multiple PDF ticket files concurrently.
 * Each file is processed independently; individual failures are captured
 * as error strings in the result summary instead of aborting the batch.
 *
 * @param onProgress - optional callback invoked after each file completes,
 *   with `(done, total)` so callers can render a progress indicator.
 */
export async function uploadTickets(
  files: File[],
  onProgress?: (done: number, total: number) => void,
): Promise<TicketUploadSummary> {
  let done = 0;
  const total = files.length;

  const results = await Promise.all(
    files.map(async (file) => {
      try {
        const result = await uploadTicket(file);
        onProgress?.(++done, total);
        return { file: file.name, ok: true as const, result };
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        onProgress?.(++done, total);
        return { file: file.name, ok: false as const, error: message };
      }
    }),
  );

  return {
    total: results.length,
    succeeded: results.filter((r) => r.ok).length,
    failed: results.filter((r) => !r.ok).length,
    items: results,
  };
}

export async function getAnalytics(): Promise<AnalyticsResult> {
  const res = await fetch(`${API_BASE}/analytics`);
  if (!res.ok) {
    throw new Error(`Analytics failed: ${res.statusText}`);
  }
  return res.json();
}
