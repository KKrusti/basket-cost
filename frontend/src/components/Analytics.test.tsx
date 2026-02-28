import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Analytics from './Analytics';
import type { AnalyticsResult } from '../types';

vi.mock('../api/products', () => ({
  getAnalytics: vi.fn(),
}));

vi.mock('./ProductImage', () => ({
  default: ({ name }: { name: string }) => <div data-testid="product-image">{name}</div>,
}));

import { getAnalytics } from '../api/products';

const mockAnalytics: AnalyticsResult = {
  mostPurchased: [
    { id: 'leche', name: 'LECHE ENTERA', purchaseCount: 5, currentPrice: 0.89 },
    { id: 'pan', name: 'PAN INTEGRAL', purchaseCount: 3, currentPrice: 1.25 },
  ],
  biggestIncreases: [
    { id: 'aceite', name: 'ACEITE OLIVA', firstPrice: 3.00, currentPrice: 6.00, increasePercent: 100.0 },
    { id: 'yogur', name: 'YOGUR NATURAL', firstPrice: 0.40, currentPrice: 0.60, increasePercent: 50.0 },
  ],
};

const emptyAnalytics: AnalyticsResult = {
  mostPurchased: [],
  biggestIncreases: [],
};

beforeEach(() => {
  vi.resetAllMocks();
});

describe('Analytics', () => {
  it('muestra un indicador de carga mientras se obtienen los datos', () => {
    vi.mocked(getAnalytics).mockReturnValue(new Promise(() => {}));
    render(<Analytics onSelectProduct={vi.fn()} />);
    expect(screen.getByRole('status', { hidden: true }) ?? document.querySelector('[aria-busy="true"]')).toBeTruthy();
  });

  it('renderiza las dos secciones de analítica con datos', async () => {
    vi.mocked(getAnalytics).mockResolvedValue(mockAnalytics);
    render(<Analytics onSelectProduct={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByText('Productos más comprados')).toBeInTheDocument();
      expect(screen.getByText('Mayor subida de precio')).toBeInTheDocument();
    });
  });

  it('muestra los productos más comprados con nombre y conteo', async () => {
    vi.mocked(getAnalytics).mockResolvedValue(mockAnalytics);
    render(<Analytics onSelectProduct={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByText('LECHE ENTERA')).toBeInTheDocument();
      expect(screen.getByText('PAN INTEGRAL')).toBeInTheDocument();
    });

    expect(screen.getByText('5 veces')).toBeInTheDocument();
    expect(screen.getByText('3 veces')).toBeInTheDocument();
  });

  it('muestra los productos con mayor subida de precio', async () => {
    vi.mocked(getAnalytics).mockResolvedValue(mockAnalytics);
    render(<Analytics onSelectProduct={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByText('ACEITE OLIVA')).toBeInTheDocument();
      expect(screen.getByText('YOGUR NATURAL')).toBeInTheDocument();
    });

    expect(screen.getByText('+100,0%')).toBeInTheDocument();
    expect(screen.getByText('+50,0%')).toBeInTheDocument();
  });

  it('muestra mensaje vacío si no hay datos de productos más comprados', async () => {
    vi.mocked(getAnalytics).mockResolvedValue(emptyAnalytics);
    render(<Analytics onSelectProduct={vi.fn()} />);

    await waitFor(() => {
      const emptyMessages = screen.getAllByText('Aún no hay datos suficientes.');
      expect(emptyMessages).toHaveLength(2);
    });
  });

  it('muestra mensaje de error si la API falla', async () => {
    vi.mocked(getAnalytics).mockRejectedValue(new Error('Network error'));
    render(<Analytics onSelectProduct={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
      expect(screen.getByRole('alert')).toHaveTextContent('Network error');
    });
  });

  it('llama a onSelectProduct al hacer clic en un producto más comprado', async () => {
    vi.mocked(getAnalytics).mockResolvedValue(mockAnalytics);
    const onSelectProduct = vi.fn();
    render(<Analytics onSelectProduct={onSelectProduct} />);

    await waitFor(() => screen.getByText('LECHE ENTERA'));

    await userEvent.click(screen.getByText('LECHE ENTERA'));
    expect(onSelectProduct).toHaveBeenCalledWith('leche');
  });

  it('llama a onSelectProduct al hacer clic en un producto con mayor subida', async () => {
    vi.mocked(getAnalytics).mockResolvedValue(mockAnalytics);
    const onSelectProduct = vi.fn();
    render(<Analytics onSelectProduct={onSelectProduct} />);

    await waitFor(() => screen.getByText('ACEITE OLIVA'));

    await userEvent.click(screen.getByText('ACEITE OLIVA'));
    expect(onSelectProduct).toHaveBeenCalledWith('aceite');
  });

  it('muestra el rango de precios (primer precio → precio actual) en subidas', async () => {
    vi.mocked(getAnalytics).mockResolvedValue(mockAnalytics);
    render(<Analytics onSelectProduct={vi.fn()} />);

    await waitFor(() => screen.getByText('ACEITE OLIVA'));

    // Los precios deben aparecer formateados en euros
    const priceTexts = screen.getAllByText(/3,00\s*€|6,00\s*€/);
    expect(priceTexts.length).toBeGreaterThanOrEqual(2);
  });

  it('muestra el rango de precios correcto para yogur', async () => {
    vi.mocked(getAnalytics).mockResolvedValue(mockAnalytics);
    render(<Analytics onSelectProduct={vi.fn()} />);

    await waitFor(() => screen.getByText('YOGUR NATURAL'));

    const priceTexts = screen.getAllByText(/0,40\s*€|0,60\s*€/);
    expect(priceTexts.length).toBeGreaterThanOrEqual(2);
  });

  it('muestra "1 vez" en singular cuando purchaseCount es 1', async () => {
    vi.mocked(getAnalytics).mockResolvedValue({
      mostPurchased: [
        { id: 'sal', name: 'SAL FINA', purchaseCount: 1, currentPrice: 0.45 },
      ],
      biggestIncreases: [],
    });
    render(<Analytics onSelectProduct={vi.fn()} />);

    await waitFor(() => screen.getByText('SAL FINA'));
    expect(screen.getByText('1 vez')).toBeInTheDocument();
  });
});
