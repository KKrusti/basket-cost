import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ProductBrowser from './ProductBrowser';
import * as productsApi from '../api/products';
import type { SearchResult } from '../types';

vi.mock('../api/products');

const mockProducts: SearchResult[] = [
  { id: '1', name: 'LECHE ENTERA HACENDADO 1L', category: 'Lácteos', currentPrice: 0.89, minPrice: 0.79, maxPrice: 0.89 },
  { id: '8', name: 'YOGUR NATURAL DANONE PACK 4', category: 'Lácteos', currentPrice: 1.79, minPrice: 1.55, maxPrice: 1.79 },
  { id: '5', name: 'ARROZ LARGO SOS 1KG', category: 'Arroces y pastas', currentPrice: 1.65, minPrice: 1.39, maxPrice: 1.65 },
  { id: '10', name: 'PLATANOS KG', category: undefined, currentPrice: 1.99, minPrice: 1.69, maxPrice: 1.99 },
];

beforeEach(() => {
  vi.clearAllMocks();
});

describe('ProductBrowser', () => {
  it('shows a loading message while fetching', () => {
    vi.mocked(productsApi.getAllProducts).mockReturnValue(new Promise(() => {}));
    render(<ProductBrowser onSelectProduct={vi.fn()} />);
    expect(screen.getByText(/cargando productos/i)).toBeInTheDocument();
  });

  it('renders a button for each product', async () => {
    vi.mocked(productsApi.getAllProducts).mockResolvedValue(mockProducts);
    render(<ProductBrowser onSelectProduct={vi.fn()} />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'LECHE ENTERA HACENDADO 1L' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'YOGUR NATURAL DANONE PACK 4' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ARROZ LARGO SOS 1KG' })).toBeInTheDocument();
  });

  it('groups products by category', async () => {
    vi.mocked(productsApi.getAllProducts).mockResolvedValue(mockProducts);
    render(<ProductBrowser onSelectProduct={vi.fn()} />);
    await waitFor(() => expect(screen.getByText('Lácteos')).toBeInTheDocument());
    expect(screen.getByText('Arroces y pastas')).toBeInTheDocument();
  });

  it('uses "Otros" for products without category', async () => {
    vi.mocked(productsApi.getAllProducts).mockResolvedValue(mockProducts);
    render(<ProductBrowser onSelectProduct={vi.fn()} />);
    await waitFor(() => expect(screen.getByText('Otros')).toBeInTheDocument());
  });

  it('calls onSelectProduct with the correct id when a button is clicked', async () => {
    vi.mocked(productsApi.getAllProducts).mockResolvedValue(mockProducts);
    const onSelect = vi.fn();
    render(<ProductBrowser onSelectProduct={onSelect} />);
    await waitFor(() => screen.getByRole('button', { name: 'LECHE ENTERA HACENDADO 1L' }));
    await userEvent.click(screen.getByRole('button', { name: 'LECHE ENTERA HACENDADO 1L' }));
    expect(onSelect).toHaveBeenCalledWith('1');
  });

  it('shows an error message when the API call fails', async () => {
    vi.mocked(productsApi.getAllProducts).mockRejectedValue(new Error('Network error'));
    render(<ProductBrowser onSelectProduct={vi.fn()} />);
    await waitFor(() =>
      expect(screen.getByText(/no se pudieron cargar los productos/i)).toBeInTheDocument()
    );
  });
});
