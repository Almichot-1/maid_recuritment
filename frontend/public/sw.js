const CACHE_VERSION = "v2";
const STATIC_CACHE = `maid-recruitment-static-${CACHE_VERSION}`;
const API_CACHE = `maid-recruitment-api-${CACHE_VERSION}`;
const FONT_CACHE = `maid-recruitment-fonts-${CACHE_VERSION}`;

const STATIC_ASSETS = [
  "/",
  "/branding/logo-light.webp",
  "/branding/logo-dark.webp",
];

const API_CACHE_CONFIG = {
  maxEntries: 100,
  maxAgeSeconds: 5 * 60,
};

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(STATIC_CACHE).then((cache) => {
      return cache.addAll(STATIC_ASSETS).catch(() => {});
    })
  );
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) => {
      return Promise.all(
        keys
          .filter((key) => {
            return (
              key !== STATIC_CACHE &&
              key !== API_CACHE &&
              key !== FONT_CACHE
            );
          })
          .map((key) => caches.delete(key))
      );
    })
  );
  self.clients.claim();
});

self.addEventListener("fetch", (event) => {
  const { request } = event;
  const url = new URL(request.url);

  if (url.origin !== self.location.origin && !url.hostname.includes("onrender.com") && !url.hostname.includes("r2.dev")) {
    return;
  }

  if (request.method !== "GET") {
    return;
  }

  if (request.url.includes("/api/v1")) {
    event.respondWith(networkFirst(request, API_CACHE));
    return;
  }

  if (
    request.destination === "font" ||
    request.url.includes(".woff") ||
    request.url.includes(".woff2")
  ) {
    event.respondWith(cacheFirst(request, FONT_CACHE));
    return;
  }

  if (
    request.destination === "image" ||
    request.url.includes("/branding/") ||
    request.url.includes("/_next/static")
  ) {
    event.respondWith(cacheFirst(request, STATIC_CACHE));
    return;
  }

  event.respondWith(networkFirst(request, STATIC_CACHE));
});

async function cacheFirst(request, cacheName) {
  const cached = await caches.match(request);
  if (cached) {
    return cached;
  }
  try {
    const response = await fetch(request);
    if (response.ok) {
      const cache = await caches.open(cacheName);
      cache.put(request, response.clone());
    }
    return response;
  } catch (error) {
    return new Response("Offline", { status: 503 });
  }
}

async function networkFirst(request, cacheName) {
  try {
    const response = await fetch(request);
    if (response.ok) {
      const cache = await caches.open(cacheName);
      cache.put(request, response.clone());
    }
    return response;
  } catch (error) {
    const cached = await caches.match(request);
    if (cached) {
      return cached;
    }

    if (request.headers.get("Accept")?.includes("application/json")) {
      return new Response(
        JSON.stringify({ error: "You are offline. Please check your internet connection." }),
        {
          status: 503,
          headers: { "Content-Type": "application/json" },
        }
      );
    }

    return new Response("You are offline. Please check your internet connection.", {
      status: 503,
    });
  }
}
