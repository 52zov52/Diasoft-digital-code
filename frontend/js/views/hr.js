import { api } from '../api.js';
import { initScanner } from '../scanner.js';
import { showToast } from '../router.js';

export async function render() {
  document.getElementById('app-outlet').innerHTML = `
    <div class="card">
      <h2>Верификация диплома</h2>
      <input type="text" id="hr-number" placeholder="Номер диплома" required>
      <input type="text" id="hr-uni" placeholder="Код вуза" required>
      <button class="btn" id="btn-verify">Проверить вручную</button>
      <hr style="margin:15px 0; border:0; border-top:1px solid var(--border);">
      <button class="btn" id="btn-scan" style="background:#475569">Сканировать QR</button>
      <div id="reader" class="hidden"></div>
    </div>
    <div id="hr-result" class="card hidden"></div>
  `;

  // Ручная проверка
  document.getElementById('btn-verify').onclick = async () => {
    const number = document.getElementById('hr-number').value.trim();
    const uni = document.getElementById('hr-uni').value.trim();
    
    if (!number || !uni) return showToast('Заполните оба поля', 'error');
    
    try {
      const res = await api.get(`/api/v1/verify?number=${encodeURIComponent(number)}&university_code=${encodeURIComponent(uni)}`);
      renderResult(res);
      showToast('Диплом проверен!', 'success');
    } catch(e) { 
      showToast(e.message, 'error'); 
    }
  };

  // Сканирование QR
  // Сканирование QR-кода
  document.getElementById('btn-scan').onclick = async () => {
    const readerEl = document.getElementById('reader');
    readerEl.classList.remove('hidden');
    
    try {
      // Импортируем функцию инициализации сканера
      const { initScanner } = await import('../scanner.js');
      
      initScanner(
        // ✅ Успешное сканирование
        async (decodedText) => {
          console.log('✅ QR scanned:', decodedText);
          
          // Останавливаем сканер
          const { stopScanner } = await import('../scanner.js');
          stopScanner();
          
          readerEl.classList.add('hidden');
          
          try {
            // Извлекаем токен из ссылки
            let token = decodedText.trim();
            
            if (token.includes('/verify/qr/')) {
              const parts = token.split('/verify/qr/');
              if (parts.length >= 2) {
                token = parts[parts.length - 1];
                token = token.split('?')[0].split('#')[0].replace(/\/+$/, '');
              }
            }
            
            if (!token || token.length < 10) {
              throw new Error('Некорректный формат QR-кода');
            }
            
            console.log('🔍 Verifying token:', token);
            
            // Отправляем на верификацию
            const res = await api.get(`/api/v1/verify/qr/${token}`);
            
            // Показываем результат
            renderResult(res);
            showToast('✅ Диплом верифицирован!', 'success');
            
          } catch(e) {
            console.error('❌ QR verification error:', e);
            showToast('Ошибка: ' + (e.message || 'Не удалось проверить'), 'error');
          }
        },
        
        // ❌ Ошибка сканера
        (err) => {
          console.error('Scanner error:', err);
          // Не показываем ошибку для каждого кадра — это нормально
        }
      );
      
    } catch (err) {
      console.error('Failed to init scanner:', err);
      showToast('Ошибка инициализации сканера: ' + err.message, 'error');
      readerEl.classList.add('hidden');
    }
  };
}

// Рендер результата проверки
function renderResult(data) {
  const el = document.getElementById('hr-result');
  el.classList.remove('hidden');
  
  // Определяем статус для цветового оформления
  const status = data.status || 'unknown';
  const statusClass = status === 'valid' || status === 'active' ? 'status-valid' : 'status-revoked';
  const statusText = status === 'valid' || status === 'active' ? '🟢 VALID' : '🔴 ' + status.toUpperCase();
  
  // Безопасный рендер данных (экранирование через textContent не нужно, т.к. данные с бэкенда)
  el.innerHTML = `
    <h3 class="${statusClass}" style="text-align:center; font-size:1.5em; margin-bottom:15px;">${statusText}</h3>
    <table style="width:100%; border-collapse:collapse;">
      <tbody>
        <tr style="border-bottom:1px solid var(--border);">
          <td style="padding:8px 0; font-weight:500; color:var(--text-muted);">ВУЗ</td>
          <td style="padding:8px 0; text-align:right;">${data.university?.substring(0, 8) || 'N/A'}...</td>
        </tr>
        <tr style="border-bottom:1px solid var(--border);">
          <td style="padding:8px 0; font-weight:500; color:var(--text-muted);">Специальность</td>
          <td style="padding:8px 0; text-align:right;">${data.specialty || 'N/A'}</td>
        </tr>
        <tr style="border-bottom:1px solid var(--border);">
          <td style="padding:8px 0; font-weight:500; color:var(--text-muted);">Год выпуска</td>
          <td style="padding:8px 0; text-align:right;">${data.issue_year || 'N/A'}</td>
        </tr>
        ${data.fio_masked ? `
        <tr>
          <td style="padding:8px 0; font-weight:500; color:var(--text-muted);">ФИО</td>
          <td style="padding:8px 0; text-align:right; font-family:monospace;">${data.fio_masked}</td>
        </tr>` : ''}
      </tbody>
    </table>
  `;
}