export interface PriceRecord {
  date: string;
  price: number;
  store?: string;
}

export interface Product {
  id: string;
  name: string;
  category?: string;
  currentPrice: number;
  priceHistory: PriceRecord[];
}

export interface SearchResult {
  id: string;
  name: string;
  category?: string;
  currentPrice: number;
  minPrice: number;
  maxPrice: number;
}
