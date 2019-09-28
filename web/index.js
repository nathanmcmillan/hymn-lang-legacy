window.onload = function () {
    const service = false
    if (service && 'serviceWorker' in navigator) {
        navigator.serviceWorker.register('/service.js').then((registration) => {
            console.log('service worker registered ', registration);
        }).catch((error) => {
            console.log('failed service worker registration', error);
        });
    }
}
