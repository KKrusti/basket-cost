import { test, expect } from '@playwright/test';
import { stubProducts, stubEmptyProducts } from './helpers';

test.beforeEach(async ({ page }) => {
  await stubProducts(page);
  await page.goto('/');
});

test('muestra el catálogo de productos en la carga inicial', async ({ page }) => {
  await expect(page.getByText('LECHE ENTERA HACENDADO 1L')).toBeVisible();
  await expect(page.getByText('PAN DE MOLDE HACENDADO')).toBeVisible();
  await expect(page.getByText('YOGUR NATURAL HACENDADO')).toBeVisible();
});

test('muestra los precios formateados en los productos', async ({ page }) => {
  await expect(page.getByText('0,89 €')).toBeVisible();
  await expect(page.getByText('1,35 €')).toBeVisible();
});

test('muestra las categorías de los productos', async ({ page }) => {
  const lacteosItems = page.getByText('Lácteos');
  await expect(lacteosItems.first()).toBeVisible();
});

test('el botón de 4 columnas cambia el layout del grid', async ({ page }) => {
  const grid = page.locator('.product-grid');
  await page.getByRole('button', { name: '4 columnas' }).click();
  await expect(grid).toHaveClass(/product-grid--4/);
});

test('el botón de 3 columnas restaura el layout por defecto', async ({ page }) => {
  const grid = page.locator('.product-grid');
  await page.getByRole('button', { name: '4 columnas' }).click();
  await page.getByRole('button', { name: '3 columnas' }).click();
  await expect(grid).toHaveClass(/product-grid--3/);
});

test('muestra estado vacío cuando la API devuelve una lista vacía', async ({ page }) => {
  await stubEmptyProducts(page);
  await page.reload();
  await expect(page.getByText(/no hay productos/i)).toBeVisible();
});

test('la paginación aparece cuando hay más productos que el tamaño de página', async ({ page }) => {
  // Generate 60 products to exceed default pageSize of 48
  const manyProducts = Array.from({ length: 60 }, (_, i) => ({
    id: `producto-${i}`,
    name: `PRODUCTO ${i}`,
    category: 'Varios',
    currentPrice: 1.0 + i * 0.01,
    imageUrl: null,
    lastPurchaseDate: '2025-09-01T00:00:00Z',
  }));
  await page.route('/api/products*', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(manyProducts),
    }),
  );
  await page.reload();
  await expect(page.getByRole('button', { name: /siguiente/i })).toBeVisible();
});

test('navega a la página siguiente al pulsar el botón siguiente', async ({ page }) => {
  const manyProducts = Array.from({ length: 60 }, (_, i) => ({
    id: `producto-${i}`,
    name: `PRODUCTO ${i}`,
    category: 'Varios',
    currentPrice: 1.0,
    imageUrl: null,
    lastPurchaseDate: '2025-09-01T00:00:00Z',
  }));
  await page.route('/api/products*', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(manyProducts),
    }),
  );
  await page.reload();
  await page.getByRole('button', { name: /siguiente/i }).click();
  await expect(page.getByRole('button', { name: /anterior/i })).toBeEnabled();
});
