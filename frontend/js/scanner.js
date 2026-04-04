let html5QrcodeScanner = null;

export function initScanner(onScanSuccess, onScanError) {
  // Если сканер уже инициализирован — очищаем
  if (html5QrcodeScanner) {
    html5QrcodeScanner.clear().catch(err => {
      console.error('Failed to clear scanner', err);
    });
  }

  // Создаём новый сканер
  html5QrcodeScanner = new Html5QrcodeScanner(
    "reader", // ID элемента в HTML
    {
      fps: 10, // кадров в секунду
      qrbox: { width: 250, height: 250 }, // размер области сканирования
      aspectRatio: 1.0,
      disableFlip: false,
      videoConstraints: {
        facingMode: "environment" // используем заднюю камеру
      }
    },
    /* verbose= */ false
  );

  // Запускаем сканер
  html5QrcodeScanner.render(onScanSuccess, onScanError || ((error) => {
    console.warn('QR Scan error:', error);
  }));
}

export function stopScanner() {
  if (html5QrcodeScanner) {
    html5QrcodeScanner.clear().catch(err => {
      console.error('Failed to stop scanner', err);
    });
    html5QrcodeScanner = null;
  }
}