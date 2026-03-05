import { test, expect } from '@playwright/test';
import { stubProducts, stubProductDetail } from './helpers';

const PRODUCT_ID = 'leche-entera-hacendado-1l';

test.beforeEach(async ({ page }) => {
  await stubProducts(page);
  await stubProductDetail(page, PRODUCT_ID);
  await page.goto('/');
  // Navigate to product detail by clicking on the product card
  await page.getByText('LECHE ENTERA HACENDADO 1L').first().click();
  await expect(page.locator('.product-detail')).toBeVisible({ timeout: 5000 });
});

test('muestra el nombre del producto en el detalle', async ({ page }) => {
  await expect(page.getByRole('heading', { name: 'LECHE ENTERA HACENDADO 1L' })).toBeVisible();
});

test('muestra la categoría del producto', async ({ page }) => {
  await expect(page.getByText('Lácteos')).toBeVisible();
});

test('muestra el precio actual formateado', async ({ page }) => {
  const priceEl = page.locator('.detail-header .price');
  await expect(priceEl).toContainText('0,89 €');
});

test('muestra el historial de precios en la tabla', async ({ page }) => {
  await expect(page.getByText('Historial de precios')).toBeVisible();
  await expect(page.getByText('Mercadona')).toBeVisible();
});

test('muestra el gráfico de precios', async ({ page }) => {
  await expect(page.getByText('Price history')).toBeVisible();
});

test('muestra la badge de variación de precio', async ({ page }) => {
  // 0.79 → 0.89 = +12.7%
  await expect(page.locator('.price-change-badge')).toBeVisible();
  await expect(page.locator('.price-change-badge')).toContainText('+12,7%');
});

test('el botón "Back to search" navega de vuelta al catálogo', async ({ page }) => {
  await page.getByRole('button', { name: /back to search/i }).click();
  await expect(page.locator('.product-detail')).not.toBeVisible();
  await expect(page.locator('.product-grid')).toBeVisible();
});

test('el logo de la app navega a la home y muestra el catálogo', async ({ page }) => {
  await page.getByRole('button', { name: 'Ir a la página principal' }).click();
  await expect(page.locator('.product-detail')).not.toBeVisible();
  await expect(page.locator('.product-grid')).toBeVisible();
});
