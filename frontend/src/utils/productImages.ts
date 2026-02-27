// Static map of product images from Mercadona CDN (by product ID matching mockdata).
// URLs use the Mercadona imgix CDN obtained from tienda.mercadona.es/api/categories/{id}/?lang=es
// Fallback emojis are used when the image fails to load or no URL is mapped.

const BASE = 'https://prod-mercadona.imgix.net/images';
const Q = '?fit=crop&h=300&w=300';

const PRODUCT_IMAGE_URLS: Record<string, string> = {
  '1':  `${BASE}/b9613b9354f8b0705f998b2201ffe443.jpg${Q}`, // Leche entera (cat 72)
  '2':  `${BASE}/bb4bc87dd5cde05303b1e222c2131fb0.jpg${Q}`, // Pan de molde integral (cat 60)
  '3':  `${BASE}/a5648e373920a10023a7ab6304eb0dc0.jpg${Q}`, // Aceite oliva virgen extra (cat 112)
  '4':  `${BASE}/bdad77c847511bc5d6fa8e5fcc533823.jpg${Q}`, // Huevos (cat 77)
  '5':  `${BASE}/0daf43fb5761b823ce83c985930c97c9.jpg${Q}`, // Arroz redondo Hacendado (cat 118)
  '6':  `${BASE}/66992393aeb518b19a12b4844845ac21.jpg${Q}`, // Pasta (cat 120)
  '7':  `${BASE}/117a2a2230b103f17b50e07a73a8fc38.jpg${Q}`, // Tomate triturado Hacendado (cat 126)
  '8':  `${BASE}/0689561ce98dba5c0b3ad43e69be0f5f.jpg${Q}`, // Yogur natural Hacendado (cat 104)
  '9':  `${BASE}/c91b2f7aa5a6c53a62e40318e876dc06.jpg${Q}`, // Pollo (cat 38)
  '10': `${BASE}/e4a37940916985bf5ca166e266580c37.jpg${Q}`, // Plátano de Canarias IGP (cat 27)
  '11': `${BASE}/12cc1ef38a5b781f364ca22e46a25ad7.jpg${Q}`, // Manzana Golden (cat 27)
  '12': `${BASE}/f6b90bdf48c92f5b451aeab2a1841b90.jpg${Q}`, // Cerveza (cat 164)
  '13': `${BASE}/1bee5993195e32109b08b2742ca1452e.jpg${Q}`, // Café molido natural Hacendado (cat 83)
  '14': `${BASE}/2114d726e660470b9eea54e72125280a.jpg${Q}`, // Papel higiénico Bosque Verde (cat 238)
  '15': `${BASE}/ce125c37c76010be23a65a2f588c076a.jpg${Q}`, // Detergente Bosque Verde (cat 226)
};

const CATEGORY_EMOJI: Record<string, string> = {
  'Lácteos':          '🥛',
  'Panadería':        '🍞',
  'Aceites':          '🫙',
  'Huevos':           '🥚',
  'Arroces y pastas': '🍝',
  'Conservas':        '🥫',
  'Carnes':           '🍗',
  'Frutas':           '🍎',
  'Bebidas':          '🍺',
  'Desayuno':         '☕',
  'Higiene':          '🧻',
  'Limpieza':         '🧴',
  'Otros':            '🛒',
};

export function getProductImageUrl(id: string): string | undefined {
  return PRODUCT_IMAGE_URLS[id];
}

export function getCategoryEmoji(category: string | undefined): string {
  return CATEGORY_EMOJI[category ?? 'Otros'] ?? '🛒';
}
