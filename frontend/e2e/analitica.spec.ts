import { test, expect } from '@playwright/test';
import { stubProducts, stubAnalytics, stubProductDetail } from './helpers';

test.beforeEach(async ({ page }) => {
  await stubProducts(page);
  await stubAnalytics(page);
  await page.goto('/');
});

test('la pestaña "Analítica" es visible', async ({ page }) => {
  await expect(page.getByRole('tab', { name: 'Analítica' })).toBeVisible();
});

test('navegar a la pestaña Analítica muestra los datos', async ({ page }) => {
  await page.getByRole('tab', { name: 'Analítica' }).click();
  await expect(page.getByText('LECHE ENTERA HACENDADO 1L')).toBeVisible({ timeout: 3000 });
  await expect(page.getByText('PAN DE MOLDE HACENDADO')).toBeVisible();
});

test('muestra el número de veces comprado', async ({ page }) => {
  await page.getByRole('tab', { name: 'Analítica' }).click();
  await expect(page.getByText(/12 veces/i)).toBeVisible({ timeout: 3000 });
});

test('muestra los productos con mayor subida de precio', async ({ page }) => {
  await page.getByRole('tab', { name: 'Analítica' }).click();
  await expect(page.getByText('YOGUR NATURAL HACENDADO')).toBeVisible({ timeout: 3000 });
  await expect(page.getByText(/28,57\s*%/)).toBeVisible();
});

test('hacer click en un producto de analítica navega al detalle', async ({ page }) => {
  await stubProductDetail(page, 'leche-entera-hacendado-1l');
  await page.getByRole('tab', { name: 'Analítica' }).click();
  await page.getByText('LECHE ENTERA HACENDADO 1L').first().click();
  await expect(page.locator('.product-detail')).toBeVisible({ timeout: 5000 });
});

test('el logo navega de vuelta a la pestaña Productos desde Analítica', async ({ page }) => {
  await page.getByRole('tab', { name: 'Analítica' }).click();
  await expect(page.getByRole('tab', { name: 'Analítica' })).toHaveAttribute('aria-selected', 'true');
  await page.getByRole('button', { name: 'Ir a la página principal' }).click();
  await expect(page.getByRole('tab', { name: 'Productos' })).toHaveAttribute('aria-selected', 'true');
});
