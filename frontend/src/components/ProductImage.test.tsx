import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ProductImage from './ProductImage';

describe('ProductImage', () => {
  it('renders an img element when a URL is mapped for the product', () => {
    const { container } = render(<ProductImage productId="1" category="Lácteos" />);
    const img = container.querySelector('img');
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute('src');
    expect(img?.getAttribute('src')).toContain('prod-mercadona.imgix.net');
  });

  it('renders the emoji fallback for unknown product IDs', () => {
    // ID 9999 has no mapped URL → should show emoji div
    const { container } = render(<ProductImage productId="9999" category="Lácteos" />);
    expect(screen.queryByRole('img')).not.toBeInTheDocument();
    expect(container.querySelector('.product-img-emoji')).toBeInTheDocument();
    expect(container.querySelector('.product-img-emoji')?.textContent).toBe('🥛');
  });

  it('applies the sm size class when size="sm"', () => {
    render(<ProductImage productId="9999" category="Lácteos" size="sm" />);
    const { container } = render(<ProductImage productId="9999" category="Lácteos" size="sm" />);
    expect(container.querySelector('.product-img-sm')).toBeInTheDocument();
  });

  it('applies the md size class when size="md"', () => {
    const { container } = render(<ProductImage productId="9999" category="Lácteos" size="md" />);
    expect(container.querySelector('.product-img-md')).toBeInTheDocument();
  });

  it('applies the md size class by default', () => {
    const { container } = render(<ProductImage productId="9999" category="Lácteos" />);
    expect(container.querySelector('.product-img-md')).toBeInTheDocument();
  });

  it('uses the Otros emoji when category is undefined', () => {
    const { container } = render(<ProductImage productId="9999" category={undefined} />);
    expect(container.querySelector('.product-img-emoji')?.textContent).toBe('🛒');
  });
});
