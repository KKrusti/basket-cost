import { useState } from 'react';
import SearchBar from './components/SearchBar';
import ProductDetail from './components/ProductDetail';
import TicketUploader from './components/TicketUploader';
import Analytics from './components/Analytics';
import type { ProductBrowserState } from './components/ProductBrowser';

type Tab = 'productos' | 'analitica';

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
  const [activeTab, setActiveTab] = useState<Tab>('productos');
  const [browserState, setBrowserState] = useState<ProductBrowserState>({
    page: 0,
    pageSize: 48,
    columns: 3,
  });

  function handleSelectProduct(id: string) {
    setSelectedProductId(id);
  }

  function handleBack() {
    setSelectedProductId(null);
  }

  function handleLogoClick() {
    setSelectedProductId(null);
    setActiveTab('productos');
    setBrowserState((prev) => ({ ...prev, page: 0 }));
  }

  return (
    <div className="app">
      <header className="app-header">
        <button
          className="app-header__logo"
          onClick={handleLogoClick}
          aria-label="Ir a la página principal"
        >
          <AppLogo />
        </button>
        <p className="app-header__subtitle">
          Consulta y compara el historial de precios de tus productos favoritos
        </p>
        <div className="app-header__actions">
          <TicketUploader />
        </div>
      </header>

      {selectedProductId ? (
        <div className="app-content">
          <ProductDetail
            productId={selectedProductId}
            onBack={handleBack}
          />
        </div>
      ) : (
        <>
          <nav className="app-tabs" role="tablist" aria-label="Secciones de la aplicación">
            <button
              role="tab"
              aria-selected={activeTab === 'productos'}
              aria-controls="tab-panel-productos"
              id="tab-productos"
              className={`app-tabs__tab${activeTab === 'productos' ? ' app-tabs__tab--active' : ''}`}
              onClick={() => setActiveTab('productos')}
            >
              Productos
            </button>
            <button
              role="tab"
              aria-selected={activeTab === 'analitica'}
              aria-controls="tab-panel-analitica"
              id="tab-analitica"
              className={`app-tabs__tab${activeTab === 'analitica' ? ' app-tabs__tab--active' : ''}`}
              onClick={() => setActiveTab('analitica')}
            >
              Analítica
            </button>
          </nav>

          <div className="app-content">
            <div
              role="tabpanel"
              id="tab-panel-productos"
              aria-labelledby="tab-productos"
              hidden={activeTab !== 'productos'}
            >
              {activeTab === 'productos' && (
                <SearchBar
                  onSelectProduct={handleSelectProduct}
                  browserState={browserState}
                  onBrowserStateChange={setBrowserState}
                />
              )}
            </div>
            <div
              role="tabpanel"
              id="tab-panel-analitica"
              aria-labelledby="tab-analitica"
              hidden={activeTab !== 'analitica'}
            >
              {activeTab === 'analitica' && (
                <Analytics onSelectProduct={handleSelectProduct} />
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
