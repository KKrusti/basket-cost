import { useState } from 'react';
import SearchBar from './components/SearchBar';
import ProductDetail from './components/ProductDetail';

export default function App() {
  const [selectedProductId, setSelectedProductId] = useState<string | null>(null);

  return (
    <div className="app">
      <header>
        <h1>Basket Cost</h1>
        <p>Search for products and view their price history</p>
      </header>

      {selectedProductId ? (
        <ProductDetail
          productId={selectedProductId}
          onBack={() => setSelectedProductId(null)}
        />
      ) : (
        <SearchBar onSelectProduct={setSelectedProductId} />
      )}
    </div>
  );
}
