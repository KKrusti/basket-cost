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
    return <div className="loading">Cargando productos...</div>;
  }

  if (error) {
    return <div className="empty-state"><p>No se pudieron cargar los productos.</p></div>;
  }

  // Group products by category
  const byCategory = products.reduce<Record<string, SearchResult[]>>((acc, product) => {
    const cat = product.category ?? 'Otros';
    if (!acc[cat]) acc[cat] = [];
    acc[cat].push(product);
    return acc;
  }, {});

  const categories = Object.keys(byCategory).sort();

  return (
    <div className="product-browser">
      {categories.map((category) => (
        <div key={category} className="browser-category">
          <h3 className="browser-category-title">{category}</h3>
          <div className="browser-product-grid">
            {byCategory[category].map((product) => (
              <button
                key={product.id}
                className="browser-product-btn"
                onClick={() => onSelectProduct(product.id)}
              >
                <ProductImage productId={product.id} category={product.category} size="sm" />
                <span className="browser-product-name">{product.name}</span>
              </button>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
