import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ProductDetail from './ProductDetail';
import * as productsApi from '../api/products';
import type { Product } from '../types';

vi.mock('../api/products');
vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  LineChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Line: () => null,
  XAxis: () => null,
  YAxis: () => null,
  CartesianGrid: () => null,
  Tooltip: () => null,
}));

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
  vi.clearAllMocks();
});

describe('ProductDetail', () => {
  it('shows "Loading product..." while loading', () => {
    vi.mocked(productsApi.getProduct).mockReturnValue(new Promise(() => {}));
    render(<ProductDetail productId="1" onBack={vi.fn()} />);
    expect(screen.getByText('Loading product...')).toBeInTheDocument();
  });

  it('shows the product name after loading', async () => {
    vi.mocked(productsApi.getProduct).mockResolvedValue(mockProduct);
    render(<ProductDetail productId="1" onBack={vi.fn()} />);
    await waitFor(() =>
      expect(screen.getByText('LECHE ENTERA HACENDADO 1L')).toBeInTheDocument()
    );
  });

  it('shows the product category', async () => {
    vi.mocked(productsApi.getProduct).mockResolvedValue(mockProduct);
    render(<ProductDetail productId="1" onBack={vi.fn()} />);
    await waitFor(() => expect(screen.getByText('Lácteos')).toBeInTheDocument());
  });

  it('shows the current price formatted in the header', async () => {
    vi.mocked(productsApi.getProduct).mockResolvedValue(mockProduct);
    render(<ProductDetail productId="1" onBack={vi.fn()} />);
    await waitFor(() => {
      const priceEl = document.querySelector('.detail-header .price');
      expect(priceEl).toHaveTextContent('0,89 €');
    });
  });

  it('shows the price history table', async () => {
    vi.mocked(productsApi.getProduct).mockResolvedValue(mockProduct);
    render(<ProductDetail productId="1" onBack={vi.fn()} />);
    await waitFor(() => screen.getByText('Price history'));
    expect(screen.getByText('Date')).toBeInTheDocument();
    expect(screen.getByText('Price')).toBeInTheDocument();
    expect(screen.getByText('Store')).toBeInTheDocument();
  });

  it('shows the stores in the table', async () => {
    vi.mocked(productsApi.getProduct).mockResolvedValue(mockProduct);
    render(<ProductDetail productId="1" onBack={vi.fn()} />);
    await waitFor(() => screen.getByText('Price history'));
    expect(screen.getAllByText('Mercadona')).toHaveLength(2);
  });

  it('shows a not found message when getProduct fails', async () => {
    vi.mocked(productsApi.getProduct).mockRejectedValue(new Error('Not found'));
    render(<ProductDetail productId="9999" onBack={vi.fn()} />);
    await waitFor(() =>
      expect(screen.getByText('Product not found')).toBeInTheDocument()
    );
  });

  it('calls onBack when the back button is pressed', async () => {
    vi.mocked(productsApi.getProduct).mockResolvedValue(mockProduct);
    const onBack = vi.fn();
    render(<ProductDetail productId="1" onBack={onBack} />);
    await waitFor(() => screen.getByText('LECHE ENTERA HACENDADO 1L'));
    await userEvent.click(screen.getByRole('button', { name: /back to search/i }));
    expect(onBack).toHaveBeenCalledOnce();
  });

  it('calls getProduct with the correct ID', async () => {
    vi.mocked(productsApi.getProduct).mockResolvedValue(mockProduct);
    render(<ProductDetail productId="1" onBack={vi.fn()} />);
    expect(productsApi.getProduct).toHaveBeenCalledWith('1');
  });

  it('shows a price-change badge when there are at least 2 price records', async () => {
    vi.mocked(productsApi.getProduct).mockResolvedValue(mockProduct);
    render(<ProductDetail productId="1" onBack={vi.fn()} />);
    await waitFor(() => screen.getByText('LECHE ENTERA HACENDADO 1L'));
    // 0.79 → 0.89 = +12.7%
    expect(screen.getByText('+12,7%')).toBeInTheDocument();
  });

  it('shows the badge in red (--up modifier) when price increased', async () => {
    vi.mocked(productsApi.getProduct).mockResolvedValue(mockProduct);
    render(<ProductDetail productId="1" onBack={vi.fn()} />);
    await waitFor(() => screen.getByText('+12,7%'));
    const badge = screen.getByText('+12,7%').closest('.price-change-badge');
    expect(badge).toHaveClass('price-change-badge--up');
  });

  it('shows the badge in green (--down modifier) when price decreased', async () => {
    const cheaperProduct: Product = {
      ...mockProduct,
      currentPrice: 0.69,
      priceHistory: [
        { date: '2025-01-15T00:00:00Z', price: 0.89, store: 'Mercadona' },
        { date: '2025-09-22T00:00:00Z', price: 0.69, store: 'Mercadona' },
      ],
    };
    vi.mocked(productsApi.getProduct).mockResolvedValue(cheaperProduct);
    render(<ProductDetail productId="1" onBack={vi.fn()} />);
    await waitFor(() => screen.getByText('LECHE ENTERA HACENDADO 1L'));
    // 0.89 → 0.69 = -22.5%
    const badge = screen.getByText('-22,5%').closest('.price-change-badge');
    expect(badge).toHaveClass('price-change-badge--down');
  });

  it('does not show the badge when there is only one price record', async () => {
    const singleRecordProduct: Product = {
      ...mockProduct,
      priceHistory: [{ date: '2025-01-15T00:00:00Z', price: 0.89, store: 'Mercadona' }],
    };
    vi.mocked(productsApi.getProduct).mockResolvedValue(singleRecordProduct);
    render(<ProductDetail productId="1" onBack={vi.fn()} />);
    await waitFor(() => screen.getByText('LECHE ENTERA HACENDADO 1L'));
    expect(document.querySelector('.price-change-badge')).not.toBeInTheDocument();
  });
});
