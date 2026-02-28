import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';
import * as productsApi from './api/products';
import type { SearchResult, Product } from './types';

vi.mock('./api/products');
vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  LineChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Line: () => null,
  XAxis: () => null,
  YAxis: () => null,
  CartesianGrid: () => null,
  Tooltip: () => null,
}));

const mockResults: SearchResult[] = [
  { id: '1', name: 'LECHE ENTERA HACENDADO 1L', category: 'Lácteos', currentPrice: 0.89, minPrice: 0.79, maxPrice: 0.89 },
];

const mockProduct: Product = {
  id: '1',
  name: 'LECHE ENTERA HACENDADO 1L',
  category: 'Lácteos',
  currentPrice: 0.89,
  priceHistory: [
    { date: '2025-01-15T00:00:00Z', price: 0.79, store: 'Mercadona' },
  ],
};

beforeEach(() => {
  vi.mocked(productsApi.getAllProducts).mockResolvedValue([]);
});

describe('App', () => {
  it('renders the application logo', () => {
    render(<App />);
    expect(screen.getByRole('img', { name: /mercaflación/i })).toBeInTheDocument();
  });

  it('shows SearchBar by default', () => {
    render(<App />);
    expect(screen.getByPlaceholderText(/search product/i)).toBeInTheDocument();
  });

  it('navigates to ProductDetail when a product is selected', async () => {
    vi.mocked(productsApi.searchProducts).mockResolvedValue(mockResults);
    vi.mocked(productsApi.getProduct).mockResolvedValue(mockProduct);
    render(<App />);
    await userEvent.type(screen.getByRole('textbox'), 'leche');
    await waitFor(() => screen.getByText('LECHE ENTERA HACENDADO 1L'));
    await userEvent.click(screen.getByText('LECHE ENTERA HACENDADO 1L'));
    await waitFor(() =>
      expect(productsApi.getProduct).toHaveBeenCalledWith('1')
    );
  });

  it('returns to SearchBar when the back button is pressed', async () => {
    vi.mocked(productsApi.searchProducts).mockResolvedValue(mockResults);
    vi.mocked(productsApi.getProduct).mockResolvedValue(mockProduct);
    render(<App />);
    await userEvent.type(screen.getByRole('textbox'), 'leche');
    await waitFor(() => screen.getByText('LECHE ENTERA HACENDADO 1L'));
    await userEvent.click(screen.getByText('LECHE ENTERA HACENDADO 1L'));
    await waitFor(() => screen.getByRole('button', { name: /back to search/i }));
    await userEvent.click(screen.getByRole('button', { name: /back to search/i }));
    expect(screen.getByPlaceholderText(/search product/i)).toBeInTheDocument();
  });
});
