export function initScanner(onSuccess, onError) {
  const html5QrCode = new Html5Qrcode("reader");
  const config = { fps: 10, qrbox: { width: 250, height: 250 }, aspectRatio: 1.0 };
  
  html5QrCode.start(
    { facingMode: "environment" },
    config,
    (decodedText) => {
      html5QrCode.stop().catch(console.warn);
      onSuccess(decodedText);
    },
    () => {} // Игнорируем промежуточные ошибки сканирования
  ).catch(err => {
    onError("Камера недоступна. Разрешите доступ в настройках браузера.");
    console.error(err);
  });
  return html5QrCode;
}