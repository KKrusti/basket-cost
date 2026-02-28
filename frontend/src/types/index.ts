export interface PriceRecord {
  date: string;
  price: number;
  store?: string;
}

export interface Product {
  id: string;
  name: string;
  category?: string;
  imageUrl?: string;
  currentPrice: number;
  priceHistory: PriceRecord[];
}

export interface SearchResult {
  id: string;
  name: string;
  category?: string;
  imageUrl?: string;
  currentPrice: number;
  minPrice: number;
  maxPrice: number;
}

export interface TicketUploadResult {
  invoiceNumber: string;
  linesImported: number;
}

export type TicketUploadItem =
  | { file: string; ok: true; result: TicketUploadResult }
  | { file: string; ok: false; error: string };

export interface TicketUploadSummary {
  total: number;
  succeeded: number;
  failed: number;
  items: TicketUploadItem[];
}
