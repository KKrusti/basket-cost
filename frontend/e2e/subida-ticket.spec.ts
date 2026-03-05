import { test, expect } from '@playwright/test';
import { stubProducts } from './helpers';
import path from 'path';
import fs from 'fs';

// Minimal valid PDF bytes (just the header magic; the backend will reject it but we
// test the frontend behaviour, intercepting the network at the route level).
const FAKE_PDF = Buffer.from('%PDF-1.4 fake content for testing');

test.beforeEach(async ({ page }) => {
  await stubProducts(page);
  await page.goto('/');
});

test('el botón de subida de ticket es visible', async ({ page }) => {
  await expect(page.getByRole('button', { name: /subir ticket/i })).toBeVisible();
});

test('muestra progreso al subir un fichero y toast de éxito', async ({ page }) => {
  await page.route('/api/tickets', (route) =>
    route.fulfill({
      status: 201,
      contentType: 'application/json',
      body: JSON.stringify({ invoiceNumber: '4144-017-284404', linesImported: 23 }),
    }),
  );

  // Write temp file
  const tmpPath = path.join('/tmp', 'ticket-test.pdf');
  fs.writeFileSync(tmpPath, FAKE_PDF);

  // Intercept file chooser
  const [fileChooser] = await Promise.all([
    page.waitForEvent('filechooser'),
    page.getByRole('button', { name: /subir ticket/i }).click(),
  ]);
  await fileChooser.setFiles(tmpPath);

  // Progress panel should appear
  await expect(page.locator('.upload-progress')).toBeVisible({ timeout: 3000 });

  // Toast should appear after completion
  await expect(page.locator('.upload-toast')).toBeVisible({ timeout: 5000 });
  await expect(page.locator('.upload-toast')).toContainText(/23/);

  fs.unlinkSync(tmpPath);
});

test('muestra toast de error cuando el servidor rechaza el ticket', async ({ page }) => {
  await page.route('/api/tickets', (route) =>
    route.fulfill({ status: 422, body: 'Unprocessable entity' }),
  );

  const tmpPath = path.join('/tmp', 'bad-ticket.pdf');
  fs.writeFileSync(tmpPath, FAKE_PDF);

  const [fileChooser] = await Promise.all([
    page.waitForEvent('filechooser'),
    page.getByRole('button', { name: /subir ticket/i }).click(),
  ]);
  await fileChooser.setFiles(tmpPath);

  await expect(page.locator('.upload-toast')).toBeVisible({ timeout: 5000 });
  await expect(page.locator('.upload-toast')).toContainText(/válido|procesar/i);

  fs.unlinkSync(tmpPath);
});

test('el toast desaparece al pulsar el botón de cerrar', async ({ page }) => {
  await page.route('/api/tickets', (route) =>
    route.fulfill({
      status: 201,
      contentType: 'application/json',
      body: JSON.stringify({ invoiceNumber: '4144-017-000001', linesImported: 5 }),
    }),
  );

  const tmpPath = path.join('/tmp', 'ticket-close.pdf');
  fs.writeFileSync(tmpPath, FAKE_PDF);

  const [fileChooser] = await Promise.all([
    page.waitForEvent('filechooser'),
    page.getByRole('button', { name: /subir ticket/i }).click(),
  ]);
  await fileChooser.setFiles(tmpPath);

  await expect(page.locator('.upload-toast')).toBeVisible({ timeout: 5000 });
  await page.getByRole('button', { name: /cerrar/i }).last().click();
  await expect(page.locator('.upload-toast')).not.toBeVisible();

  fs.unlinkSync(tmpPath);
});
