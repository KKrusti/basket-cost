import { useState, useEffect, useCallback } from 'react';
import type { SearchResult } from '../types';
import { searchProducts } from '../api/products';
import ProductBrowser from './ProductBrowser';
import ProductImage from './ProductImage';

interface SearchBarProps {
  onSelectProduct: (id: string) => void;
}

export default function SearchBar({ onSelectProduct }: SearchBarProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  const doSearch = useCallback(async (q: string) => {
    if (q.trim().length === 0) {
      setResults([]);
      setSearched(false);
      return;
    }
    setLoading(true);
    try {
      const data = await searchProducts(q);
      setResults(data);
      setSearched(true);
    } catch (err) {
      console.error('Search error:', err);
      setResults([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => {
      doSearch(query);
    }, 300);
    return () => clearTimeout(timer);
  }, [query, doSearch]);

  const formatPrice = (price: number) =>
    price.toFixed(2).replace('.', ',') + ' \u20AC';

  return (
    <div>
      <div className="search-container">
        <input
          type="text"
          className="search-input"
          placeholder="Search product... (e.g. leche, aceite, pan)"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          autoFocus
        />
      </div>

      {loading && <div className="loading">Searching...</div>}

      {!loading && searched && results.length === 0 && (
        <div className="empty-state">
          <p>No products found for &quot;{query}&quot;</p>
        </div>
      )}

      {!loading && results.length > 0 && (
        <div className="product-list">
          {results.map((product) => (
            <div
              key={product.id}
              className="product-card"
              onClick={() => onSelectProduct(product.id)}
            >
              <ProductImage productId={product.id} category={product.category} size="md" />
              <div className="product-card-info">
                <h3>{product.name}</h3>
                {product.category && (
                  <span className="category">{product.category}</span>
                )}
              </div>
              <div className="product-card-price">
                <div className="current">{formatPrice(product.currentPrice)}</div>
                <div className="range">
                  {formatPrice(product.minPrice)} - {formatPrice(product.maxPrice)}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {!loading && !searched && (
        <ProductBrowser onSelectProduct={onSelectProduct} />
      )}
    </div>
  );
}
