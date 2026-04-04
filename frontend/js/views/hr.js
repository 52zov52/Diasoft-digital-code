import { api } from '../api.js';
import { initScanner } from '../scanner.js';
import { showToast } from '../router.js';

export async function render() {
  document.getElementById('app-outlet').innerHTML = `
    <div class="card">
      <h2>🔍 Верификация диплома</h2>
      <input type="text" id="hr-number" placeholder="Номер диплома" required>
      <input type="text" id="hr-uni" placeholder="Код вуза" required>
      <button class="btn" id="btn-verify">Проверить вручную</button>
      <hr style="margin:15px 0; border:0; border-top:1px solid var(--border);">
      <button class="btn" id="btn-scan" style="background:#475569">📷 Сканировать QR</button>
      <div id="reader"></div>
    </div>
    <div id="hr-result" class="card hidden"></div>
  `;

  document.getElementById('btn-verify').onclick = async () => {
    const number = document.getElementById('hr-number').value.trim();
    const uni = document.getElementById('hr-uni').value.trim();
    if (!number || !uni) return showToast('Заполните оба поля', 'error');
    try {
      const res = await api.get(`/api/v1/verify?number=${encodeURIComponent(number)}&university_code=${encodeURIComponent(uni)}`);
      renderResult(res);
    } catch(e) { showToast(e.message, 'error'); }
  };

  document.getElementById('btn-scan').onclick = () => {
    document.getElementById('reader').classList.remove('hidden');
    initScanner(
      async (decodedText) => {
        document.getElementById('reader').classList.add('hidden');
        const token = decodedText.split('/').pop();
        try {
          const res = await api.get(`/api/v1/verify/qr/${token}`);
          renderResult(res);
        } catch(e) { showToast(e.message, 'error'); }
      },
      (err) => { showToast(err, 'error'); document.getElementById('reader').classList.add('hidden'); }
    );
  };
}

function renderResult(data) {
  const el = document.getElementById('hr-result');
  el.classList.remove('hidden');
  const statusClass = data.status === 'valid' ? 'status-valid' : 'status-revoked';
  // Безопасный рендер без innerHTML для PII
  el.innerHTML = `
    <h3 class="${statusClass}">Статус: ${data.status?.toUpperCase()}</h3>
    <table>
      <tr><td>ВУЗ</td><td>${data.university || 'N/A'}</td></tr>
      <tr><td>Специальность</td><td>${data.specialty || 'N/A'}</td></tr>
      <tr><td>Год выпуска</td><td>${data.issue_year || 'N/A'}</td></tr>
      ${data.fio_masked ? `<tr><td>ФИО</td><td>${data.fio_masked}</td></tr>` : ''}
    </table>
  `;
}