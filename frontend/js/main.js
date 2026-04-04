// frontend/js/main.js
// Точка входа приложения

import './routes.js';

// Глобальный обработчик ошибок (для отладки)
window.addEventListener('error', (e) => {
  console.error('🔴 Global error:', e.message, e.filename, e.lineno);
});

window.addEventListener('unhandledrejection', (e) => {
  console.error('🔴 Unhandled promise:', e.reason);
});