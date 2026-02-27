import { useState } from 'react';
import SearchBar from './components/SearchBar';
import ProductDetail from './components/ProductDetail';

function BasketIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      style={{ color: 'var(--color-primary)' }}
    >
      <path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z" />
      <line x1="3" y1="6" x2="21" y2="6" />
      <path d="M16 10a4 4 0 0 1-8 0" />
    </svg>
  );
}

export default function App() {
  const [selectedProductId, setSelectedProductId] = useState<string | null>(null);

  return (
    <div className="app">
      <header className="app-header">
        <div className="app-header__logo">
          <BasketIcon />
          <h1 className="app-header__title">Basket Cost</h1>
        </div>
        <p className="app-header__subtitle">
          Consulta y compara el historial de precios de tus productos favoritos
        </p>
      </header>

      <div className="app-content">
        {selectedProductId ? (
          <ProductDetail
            productId={selectedProductId}
            onBack={() => setSelectedProductId(null)}
          />
        ) : (
          <SearchBar onSelectProduct={setSelectedProductId} />
        )}
      </div>
    </div>
  );
}
