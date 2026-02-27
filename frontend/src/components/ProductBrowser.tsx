import { useState, useEffect } from 'react';
import type { SearchResult } from '../types';
import { getAllProducts } from '../api/products';
import ProductImage from './ProductImage';

interface ProductBrowserProps {
  onSelectProduct: (id: string) => void;
}

export default function ProductBrowser({ onSelectProduct }: ProductBrowserProps) {
  const [products, setProducts] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  useEffect(() => {
    let cancelled = false;
    getAllProducts()
      .then((data) => {
        if (!cancelled) {
          setProducts(data);
          setLoading(false);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setError(true);
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  if (loading) {
    return (
      <div>
        <div className="loading">Cargando productos...</div>
        <div className="browser-product-grid" aria-hidden="true">
          {[1, 2, 3, 4, 5, 6].map((n) => (
            <div key={n} className="skeleton" style={{ height: '100px' }} />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="empty-state">
        <p>No se pudieron cargar los productos.</p>
      </div>
    );
  }

  // Group products by category
  const byCategory = products.reduce<Record<string, SearchResult[]>>((acc, product) => {
    const cat = product.category ?? 'Otros';
    if (!acc[cat]) acc[cat] = [];
    acc[cat].push(product);
    return acc;
  }, {});

  const categories = Object.keys(byCategory).sort();

  // Flatten into a single ordered list: [category header, product, product, ...]
  const items: Array<{ type: 'header'; label: string } | { type: 'product'; product: SearchResult }> = [];
  for (const category of categories) {
    items.push({ type: 'header', label: category });
    for (const product of byCategory[category]) {
      items.push({ type: 'product', product });
    }
  }

  return (
    <div className="product-browser">
      <p className="product-browser__intro">Explorar catálogo</p>
      <div className="browser-flow">
        {items.map((item) =>
          item.type === 'header' ? (
            <div key={`cat-${item.label}`} className="browser-category-label">
              {item.label}
            </div>
          ) : (
            <button
              key={item.product.id}
              className="browser-product-btn"
              onClick={() => onSelectProduct(item.product.id)}
              aria-label={item.product.name}
            >
              <ProductImage productId={item.product.id} category={item.product.category} size="sm" />
              <span className="browser-product-name">{item.product.name}</span>
            </button>
          )
        )}
      </div>
    </div>
  );
}
