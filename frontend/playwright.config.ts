import { defineConfig, devices } from '@playwright/test';

// Required shared libraries not installed system-wide; extracted from deb packages.
const PLAYWRIGHT_LIBS = '/home/carlos/.local/lib/playwright-libs';
const ldLibraryPath = process.env.LD_LIBRARY_PATH
  ? `${PLAYWRIGHT_LIBS}:${process.env.LD_LIBRARY_PATH}`
  : PLAYWRIGHT_LIBS;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: [['list'], ['html', { open: 'never' }]],

  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    launchOptions: {
      env: {
        LD_LIBRARY_PATH: ldLibraryPath,
      },
    },
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'mobile-poco-x6-pro',
      use: {
        ...devices['Pixel 5'],
        viewport: { width: 393, height: 852 },
        userAgent:
          'Mozilla/5.0 (Linux; Android 14; Poco X6 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36',
        isMobile: true,
        hasTouch: true,
      },
    },
  ],

  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
