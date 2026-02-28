import { useState } from 'react';
import SearchBar from './components/SearchBar';
import ProductDetail from './components/ProductDetail';
import TicketUploader from './components/TicketUploader';
import type { ProductBrowserState } from './components/ProductBrowser';

function AppLogo() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 400 100"
      aria-label="Mercaflación"
      role="img"
      className="app-header__logo-svg"
    >
      <rect x="50" y="10" width="300" height="80" rx="40" ry="40" fill="#00A859" stroke="#FFFFFF" strokeWidth="4"/>
      <text x="200" y="60" fontFamily="Arial, sans-serif" fontSize="32" fontWeight="bold" fill="#FFFFFF" textAnchor="middle" letterSpacing="1">MERCAFLACION</text>
    </svg>
  );
}

export default function App() {
  const [selectedProductId, setSelectedProductId] = useState<string | null>(null);
  const [browserState, setBrowserState] = useState<ProductBrowserState>({
    page: 0,
    pageSize: 48,
    columns: 3,
  });

  return (
    <div className="app">
      <header className="app-header">
        <div className="app-header__logo">
          <AppLogo />
        </div>
        <p className="app-header__subtitle">
          Consulta y compara el historial de precios de tus productos favoritos
        </p>
        <div className="app-header__actions">
          <TicketUploader />
        </div>
      </header>

      <div className="app-content">
        {selectedProductId ? (
          <ProductDetail
            productId={selectedProductId}
            onBack={() => setSelectedProductId(null)}
          />
        ) : (
          <SearchBar
            onSelectProduct={setSelectedProductId}
            browserState={browserState}
            onBrowserStateChange={setBrowserState}
          />
        )}
      </div>
    </div>
  );
}
