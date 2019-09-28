const cache_name = 'dev'

self.addEventListener('install', function (event) {
    const files_to_cache = [
        'roboto.ttf',
    ]
    event.waitUntil(caches.open(cache_name).then(function (cache) {
        console.log('cache', cache_name, 'opened')
        return cache.addAll(files_to_cache)
    }).then(function () {
        return self.skipWaiting()
    }))
})

self.addEventListener('activate', function (event) {
    console.log('cache', cache_name, 'activate')
    event.waitUntil(caches.keys().then(function (keyList) {
        return Promise.all(keyList.map(function (key) {
            if (key !== cache_name) {
                console.log(cache_name, 'removing old cache', key)
                return caches.delete(key)
            }
        }))
    }))
    return self.clients.claim()
})

self.addEventListener('fetch', function (event) {
    event.respondWith(caches.match(event.request).then((response) => {
        return response || fetch(event.request)
    }))
})
