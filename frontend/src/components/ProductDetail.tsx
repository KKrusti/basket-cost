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

interface ProductDetailProps {
  productId: string;
  onBack: () => void;
}

interface ChartDataPoint {
  date: string;
  price: number;
  store: string;
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
          &larr; Back to search
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
      <button className="back-btn" onClick={onBack}>
        &larr; Back to search
      </button>

      <div className="detail-header">
        <h2>{product.name}</h2>
        {product.category && <span className="category">{product.category}</span>}
        <div className="price">{formatPrice(product.currentPrice)}</div>
      </div>

      <div className="chart-container">
        <h3>Price history</h3>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={chartData} margin={{ top: 5, right: 20, bottom: 5, left: 10 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#eee" />
            <XAxis
              dataKey="date"
              tickFormatter={formatDateShort}
              tick={{ fontSize: 12 }}
            />
            <YAxis
              domain={[yMin, yMax]}
              tickFormatter={(v: number) => formatPrice(v)}
              tick={{ fontSize: 12 }}
              width={70}
            />
            <Tooltip
              formatter={(value: number) => [formatPrice(value), 'Price']}
              labelFormatter={(label: string) => formatDateFull(label)}
            />
            <Line
              type="monotone"
              dataKey="price"
              stroke="#2d6a4f"
              strokeWidth={2}
              dot={{ r: 4, fill: '#2d6a4f' }}
              activeDot={{ r: 6 }}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>

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
              <td>{formatPrice(record.price)}</td>
              <td>{record.store || '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
