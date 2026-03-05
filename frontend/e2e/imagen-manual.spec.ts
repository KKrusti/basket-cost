import { test, expect } from '@playwright/test';
import { stubProducts, stubProductDetail } from './helpers';

const PRODUCT_ID = 'leche-entera-hacendado-1l';
const NEW_IMAGE_URL = 'https://prod.mercadona.com/images/leche.jpg';

test.beforeEach(async ({ page }) => {
  await stubProducts(page);
  await stubProductDetail(page, PRODUCT_ID);
  await page.goto('/');
  await page.getByText('LECHE ENTERA HACENDADO 1L').first().click();
  await expect(page.locator('.product-detail')).toBeVisible({ timeout: 5000 });
});

test('el botón de editar imagen es visible en el detalle del producto', async ({ page }) => {
  await expect(
    page.getByRole('button', { name: 'Cambiar imagen del producto' }),
  ).toBeVisible();
});

test('al pulsar el botón aparece el input de URL', async ({ page }) => {
  await page.getByRole('button', { name: 'Cambiar imagen del producto' }).click();
  await expect(page.getByLabel('URL de imagen del producto')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Guardar imagen' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Cancelar' })).toBeVisible();
});

test('cancelar oculta el input y vuelve a mostrar el botón de editar', async ({ page }) => {
  await page.getByRole('button', { name: 'Cambiar imagen del producto' }).click();
  await page.getByRole('button', { name: 'Cancelar' }).click();
  await expect(page.getByLabel('URL de imagen del producto')).not.toBeVisible();
  await expect(
    page.getByRole('button', { name: 'Cambiar imagen del producto' }),
  ).toBeVisible();
});

test('guardar llama al PATCH y cierra el input', async ({ page }) => {
  let patchCalled = false;
  await page.route(`/api/products/${PRODUCT_ID}/image`, (route) => {
    patchCalled = true;
    route.fulfill({ status: 200, contentType: 'application/json', body: '{}' });
  });

  await page.getByRole('button', { name: 'Cambiar imagen del producto' }).click();
  await page.getByLabel('URL de imagen del producto').fill(NEW_IMAGE_URL);
  await page.getByRole('button', { name: 'Guardar imagen' }).click();

  await expect(page.getByLabel('URL de imagen del producto')).not.toBeVisible({ timeout: 3000 });
  expect(patchCalled).toBe(true);
});

test('muestra error si el servidor devuelve un error al guardar', async ({ page }) => {
  await page.route(`/api/products/${PRODUCT_ID}/image`, (route) =>
    route.fulfill({ status: 500, body: 'Internal Server Error' }),
  );

  await page.getByRole('button', { name: 'Cambiar imagen del producto' }).click();
  await page.getByLabel('URL de imagen del producto').fill(NEW_IMAGE_URL);
  await page.getByRole('button', { name: 'Guardar imagen' }).click();

  await expect(page.getByRole('alert')).toContainText(/No se pudo guardar/i);
});

test('muestra validación si se intenta guardar con URL vacía', async ({ page }) => {
  await page.getByRole('button', { name: 'Cambiar imagen del producto' }).click();
  await page.getByRole('button', { name: 'Guardar imagen' }).click();
  await expect(page.getByRole('alert')).toContainText(/URL/i);
});
