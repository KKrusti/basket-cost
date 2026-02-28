import { useState } from 'react';
import { getProductImageUrl, getCategoryEmoji } from '../utils/productImages';

interface ProductImageProps {
  productId: string;
  imageUrl?: string;
  category: string | undefined;
  size?: 'sm' | 'md';
}

export default function ProductImage({ productId, imageUrl, category, size = 'md' }: ProductImageProps) {
  // Priority: imageUrl from backend → static map fallback → category emoji
  const url = imageUrl || getProductImageUrl(productId);
  const [imgFailed, setImgFailed] = useState(false);

  const showEmoji = !url || imgFailed;
  const sizeClass = size === 'sm' ? 'product-img-sm' : 'product-img-md';

  if (showEmoji) {
    return (
      <div className={`product-img-emoji ${sizeClass}`} aria-hidden="true">
        {getCategoryEmoji(category)}
      </div>
    );
  }

  return (
    <img
      src={url}
      alt=""
      className={`product-img ${sizeClass}`}
      onError={() => setImgFailed(true)}
      loading="lazy"
    />
  );
}
