import { test, expect } from '@playwright/test';
import {
  loginViaStorage,
  stubEmptyProducts,
  stubHouseholdMembers,
  stubHouseholdEmpty,
  stubCreateInvitation,
  stubAcceptInvitation,
} from './helpers';

test.beforeEach(async ({ page }) => {
  await loginViaStorage(page);
  await stubEmptyProducts(page);
});

// ---------------------------------------------------------------------------
// Household section in user menu
// ---------------------------------------------------------------------------

test('muestra la sección "Unidad familiar" al abrir el menú de usuario', async ({ page }) => {
  await stubHouseholdEmpty(page);
  await page.goto('/');

  await page.getByRole('button', { name: /testuser/i }).click();
  await expect(page.getByText('Unidad familiar')).toBeVisible();
});

test('muestra el botón "Invitar conviviente" cuando el usuario no tiene unidad familiar', async ({ page }) => {
  await stubHouseholdEmpty(page);
  await page.goto('/');

  await page.getByRole('button', { name: /testuser/i }).click();
  await expect(page.getByRole('button', { name: /invitar conviviente/i })).toBeVisible();
});

test('muestra los miembros de la unidad familiar', async ({ page }) => {
  await stubHouseholdMembers(page);
  await page.goto('/');

  await page.getByRole('button', { name: /testuser/i }).click();
  await expect(page.getByRole('list', { name: /miembros de la unidad familiar/i })).toBeVisible();
  await expect(page.getByText('alice')).toBeVisible();
});

test('marca al usuario actual con "(tú)" en la lista de miembros', async ({ page }) => {
  await stubHouseholdMembers(page);
  await page.goto('/');

  await page.getByRole('button', { name: /testuser/i }).click();
  await expect(page.getByText(/\(tú\)/)).toBeVisible();
});

test('muestra el botón "Abandonar unidad" cuando el usuario está en una unidad familiar', async ({ page }) => {
  await stubHouseholdMembers(page);
  await page.goto('/');

  await page.getByRole('button', { name: /testuser/i }).click();
  await expect(page.getByRole('button', { name: /abandonar unidad familiar/i })).toBeVisible();
});

// ---------------------------------------------------------------------------
// Create invitation link
// ---------------------------------------------------------------------------

test('muestra el enlace de invitación al pulsar "Invitar conviviente"', async ({ page }) => {
  await stubHouseholdEmpty(page);
  await stubCreateInvitation(page, 'tok-abc');
  await page.goto('/');

  await page.getByRole('button', { name: /testuser/i }).click();
  await page.getByRole('button', { name: /invitar conviviente/i }).click();

  const input = page.getByRole('textbox', { name: /enlace de invitación/i });
  await expect(input).toBeVisible();
  await expect(input).toHaveValue(/invite=tok-abc/);
  await expect(page.getByText(/válido durante 24 horas/i)).toBeVisible();
});

test('muestra el botón de copiar después de generar el enlace', async ({ page }) => {
  await stubHouseholdEmpty(page);
  await stubCreateInvitation(page, 'tok-xyz');
  await page.goto('/');

  await page.getByRole('button', { name: /testuser/i }).click();
  await page.getByRole('button', { name: /invitar conviviente/i }).click();

  await expect(page.getByRole('button', { name: /copiar enlace/i })).toBeVisible();
});

test('muestra error cuando falla la creación de invitación', async ({ page }) => {
  await stubHouseholdEmpty(page);
  await page.route('/api/household/invite', (route) =>
    route.fulfill({ status: 500, body: 'Error' }),
  );
  await page.goto('/');

  await page.getByRole('button', { name: /testuser/i }).click();
  await page.getByRole('button', { name: /invitar conviviente/i }).click();

  await expect(page.getByRole('alert')).toContainText(/no se pudo crear la invitación/i);
});

// ---------------------------------------------------------------------------
// Leave household
// ---------------------------------------------------------------------------

test('cierra el menú al abandonar la unidad familiar', async ({ page }) => {
  await stubHouseholdMembers(page);
  await page.goto('/');

  await page.getByRole('button', { name: /testuser/i }).click();
  await page.getByRole('button', { name: /abandonar unidad familiar/i }).click();

  // The onLeft callback closes the dropdown
  await expect(page.getByRole('menu')).not.toBeVisible();
});

test('muestra error cuando falla al abandonar la unidad familiar', async ({ page }) => {
  // Stub GET to return members, DELETE to fail
  await page.route('/api/household', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ members: [{ id: 1, username: 'testuser' }] }),
      });
    }
    if (route.request().method() === 'DELETE') {
      return route.fulfill({ status: 500, body: 'Error' });
    }
    return route.fallback();
  });
  await page.goto('/');

  await page.getByRole('button', { name: /testuser/i }).click();
  await page.getByRole('button', { name: /abandonar unidad familiar/i }).click();

  await expect(page.getByRole('alert')).toContainText(/no se pudo abandonar/i);
});

// ---------------------------------------------------------------------------
// Accept invitation flow (/?invite=TOKEN)
// ---------------------------------------------------------------------------

test('muestra el modal de invitación cuando la URL contiene ?invite=TOKEN (usuario autenticado)', async ({ page }) => {
  await stubAcceptInvitation(page);
  await page.goto('/?invite=tok-invite-123');

  await expect(page.getByRole('dialog')).toBeVisible();
  await expect(page.getByRole('heading', { name: /unidad familiar/i })).toBeVisible();
  await expect(page.getByRole('button', { name: /aceptar invitación/i })).toBeVisible();
  await expect(page.getByRole('button', { name: /cancelar/i })).toBeVisible();
});

test('cierra el modal al pulsar Cancelar', async ({ page }) => {
  await stubAcceptInvitation(page);
  await page.goto('/?invite=tok-invite-123');

  await page.getByRole('button', { name: /cancelar/i }).click();

  await expect(page.getByRole('dialog')).not.toBeVisible();
});

test('llama a POST /api/household/accept al aceptar la invitación', async ({ page }) => {
  let acceptCalled = false;
  await page.route('/api/household/accept*', (route) => {
    acceptCalled = true;
    return route.fulfill({ status: 200, contentType: 'application/json', body: '{}' });
  });

  await page.goto('/?invite=tok-invite-123');
  await page.getByRole('button', { name: /aceptar invitación/i }).click();

  await expect.poll(() => acceptCalled).toBe(true);
});

test('muestra error en el modal cuando la invitación ha expirado', async ({ page }) => {
  await page.route('/api/household/accept*', (route) =>
    route.fulfill({ status: 404, body: 'Not Found' }),
  );
  await page.goto('/?invite=tok-expired');

  await page.getByRole('button', { name: /aceptar invitación/i }).click();

  await expect(page.getByRole('alert')).toContainText(/la invitación no existe o ha expirado/i);
});

// ---------------------------------------------------------------------------
// Unauthenticated user does NOT see the accept invite modal
// ---------------------------------------------------------------------------

test.describe('sin sesión iniciada', () => {
  // The outer beforeEach registers an addInitScript that sets auth in localStorage.
  // We counteract it by registering a second addInitScript that removes the same key.
  // addInitScripts execute in registration order, so the removal runs after the set.
  test.beforeEach(async ({ page }) => {
    await page.addInitScript((key) => {
      localStorage.removeItem(key);
    }, 'mercaflacion_auth');
  });

  test('no muestra el modal de invitación si el usuario no está autenticado', async ({ page }) => {
    await page.goto('/?invite=tok-invite-123');

    await expect(page.getByRole('dialog')).not.toBeVisible();
  });
});
