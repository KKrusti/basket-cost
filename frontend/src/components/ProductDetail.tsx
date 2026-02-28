import { useState, useEffect } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import type { Product } from '../types';
import { getProduct } from '../api/products';
import ProductImage from './ProductImage';

interface ProductDetailProps {
  productId: string;
  onBack: () => void;
}

interface ChartDataPoint {
  date: string;
  price: number;
  store: string;
}

function BackArrowIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <line x1="19" y1="12" x2="5" y2="12" />
      <polyline points="12 19 5 12 12 5" />
    </svg>
  );
}

export default function ProductDetail({ productId, onBack }: ProductDetailProps) {
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    getProduct(productId)
      .then((data) => {
        if (!cancelled) setProduct(data);
      })
      .catch((err) => console.error('Error loading product:', err))
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [productId]);

  if (loading) {
    return <div className="loading">Loading product...</div>;
  }

  if (!product) {
    return (
      <div className="empty-state">
        <p>Product not found</p>
        <button className="back-btn" onClick={onBack}>
          <BackArrowIcon />
          Back to search
        </button>
      </div>
    );
  }

  const formatPrice = (price: number) =>
    price.toFixed(2).replace('.', ',') + ' \u20AC';

  const formatDateShort = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleDateString('es-ES', { month: 'short', year: '2-digit' });
  };

  const formatDateFull = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleDateString('es-ES', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    });
  };

  const chartData: ChartDataPoint[] = product.priceHistory.map((record) => ({
    date: record.date,
    price: record.price,
    store: record.store || '',
  }));

  const prices = product.priceHistory.map((r) => r.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const yMin = Math.floor((minPrice - 0.1) * 10) / 10;
  const yMax = Math.ceil((maxPrice + 0.1) * 10) / 10;

  return (
    <div className="product-detail">
      <button className="back-btn" onClick={onBack} aria-label="Back to search">
        <BackArrowIcon />
        Back to search
      </button>

      <div className="detail-header">
        <ProductImage
          productId={product.id}
          imageUrl={product.imageUrl}
          category={product.category}
          size="lg"
        />
        <div className="detail-header__info">
          <h2>{product.name}</h2>
          {product.category && (
            <span className="category">{product.category}</span>
          )}
        </div>
        <div className="detail-header__price">
          <div className="price">{formatPrice(product.currentPrice)}</div>
          <div className="detail-header__price-label">precio actual</div>
        </div>
      </div>

      <hr className="detail-divider" />

      <div className="chart-container">
        <h3 className="chart-container__title">Price history</h3>
        <ResponsiveContainer width="100%" height={280}>
          <LineChart data={chartData} margin={{ top: 5, right: 20, bottom: 5, left: 10 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
            <XAxis
              dataKey="date"
              tickFormatter={formatDateShort}
              tick={{ fontSize: 12, fill: 'var(--color-text-muted)' }}
              axisLine={{ stroke: 'var(--color-border)' }}
              tickLine={false}
            />
            <YAxis
              domain={[yMin, yMax]}
              tickFormatter={(v: number) => formatPrice(v)}
              tick={{ fontSize: 12, fill: 'var(--color-text-muted)' }}
              width={72}
              axisLine={false}
              tickLine={false}
            />
            <Tooltip
              formatter={(value: number) => [formatPrice(value), 'Price']}
              labelFormatter={(label: string) => formatDateFull(label)}
              contentStyle={{
                borderRadius: '10px',
                border: '1.5px solid var(--color-border)',
                boxShadow: 'var(--shadow-md)',
                fontSize: '0.875rem',
              }}
            />
            <Line
              type="monotone"
              dataKey="price"
              stroke="var(--color-primary)"
              strokeWidth={2.5}
              dot={{ r: 4, fill: 'var(--color-primary)', strokeWidth: 0 }}
              activeDot={{ r: 6, fill: 'var(--color-primary-dark)', strokeWidth: 0 }}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <hr className="detail-divider" />

      <div>
        <h3 className="price-table-section__title">Historial de precios</h3>
        <div className="price-table-wrapper">
          <table className="price-table">
            <thead>
              <tr>
                <th>Date</th>
                <th>Price</th>
                <th>Store</th>
              </tr>
            </thead>
            <tbody>
              {[...product.priceHistory].reverse().map((record, i) => (
                <tr key={i}>
                  <td>{formatDateFull(record.date)}</td>
                  <td className="price-cell">{formatPrice(record.price)}</td>
                  <td className="store-cell">{record.store || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
