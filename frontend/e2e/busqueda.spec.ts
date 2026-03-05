import { test, expect } from '@playwright/test';
import { stubProducts } from './helpers';

test.beforeEach(async ({ page }) => {
  await stubProducts(page);
  await page.goto('/');
});

test('la barra de búsqueda está visible', async ({ page }) => {
  await expect(page.getByPlaceholder(/buscar producto/i)).toBeVisible();
});

test('escribir en el buscador oculta el catálogo y muestra resultados', async ({ page }) => {
  // Stub the search response
  await page.route('/api/products?q=leche', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        {
          id: 'leche-entera-hacendado-1l',
          name: 'LECHE ENTERA HACENDADO 1L',
          category: 'Lácteos',
          currentPrice: 0.89,
          imageUrl: null,
        },
      ]),
    }),
  );
  const searchInput = page.getByPlaceholder(/buscar producto/i);
  await searchInput.fill('leche');
  // Wait for debounce + result
  await expect(page.getByText('LECHE ENTERA HACENDADO 1L')).toBeVisible({ timeout: 2000 });
  // ProductBrowser should be hidden
  await expect(page.locator('.product-grid')).not.toBeVisible();
});

test('muestra "Sin resultados" cuando la búsqueda no encuentra nada', async ({ page }) => {
  await page.route('/api/products?q=xyz', (route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: '[]' }),
  );
  await page.getByPlaceholder(/buscar producto/i).fill('xyz');
  await expect(page.getByText(/sin resultados/i)).toBeVisible({ timeout: 2000 });
});

test('limpiar el buscador vuelve a mostrar el catálogo', async ({ page }) => {
  const searchInput = page.getByPlaceholder(/buscar producto/i);
  await searchInput.fill('leche');
  await searchInput.clear();
  await expect(page.locator('.product-grid')).toBeVisible({ timeout: 2000 });
});

test('hacer click en un resultado navega al detalle del producto', async ({ page }) => {
  await page.route('/api/products?q=leche', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        {
          id: 'leche-entera-hacendado-1l',
          name: 'LECHE ENTERA HACENDADO 1L',
          category: 'Lácteos',
          currentPrice: 0.89,
          imageUrl: null,
        },
      ]),
    }),
  );
  await page.route('/api/products/leche-entera-hacendado-1l', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        id: 'leche-entera-hacendado-1l',
        name: 'LECHE ENTERA HACENDADO 1L',
        category: 'Lácteos',
        currentPrice: 0.89,
        imageUrl: null,
        priceHistory: [
          { date: '2025-09-22T00:00:00Z', price: 0.89, store: 'Mercadona' },
        ],
      }),
    }),
  );
  await page.getByPlaceholder(/buscar producto/i).fill('leche');
  await page.getByText('LECHE ENTERA HACENDADO 1L').first().click();
  await expect(page.locator('.product-detail')).toBeVisible({ timeout: 3000 });
});

test('el botón volver desde detalle regresa al estado de búsqueda', async ({ page }) => {
  await page.route('/api/products?q=leche', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        {
          id: 'leche-entera-hacendado-1l',
          name: 'LECHE ENTERA HACENDADO 1L',
          category: 'Lácteos',
          currentPrice: 0.89,
          imageUrl: null,
        },
      ]),
    }),
  );
  await page.route('/api/products/leche-entera-hacendado-1l', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        id: 'leche-entera-hacendado-1l',
        name: 'LECHE ENTERA HACENDADO 1L',
        category: 'Lácteos',
        currentPrice: 0.89,
        imageUrl: null,
        priceHistory: [{ date: '2025-09-22T00:00:00Z', price: 0.89, store: 'Mercadona' }],
      }),
    }),
  );
  await page.getByPlaceholder(/buscar producto/i).fill('leche');
  await page.getByText('LECHE ENTERA HACENDADO 1L').first().click();
  await expect(page.locator('.product-detail')).toBeVisible({ timeout: 3000 });
  await page.getByRole('button', { name: /back to search/i }).click();
  await expect(page.locator('.product-detail')).not.toBeVisible();
});
