import type { SearchResult, Product, TicketUploadResult } from '../types';

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
