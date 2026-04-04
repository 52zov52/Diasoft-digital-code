import { api } from '../api.js';
import { showToast } from '../router.js';

export async function render() {
  document.getElementById('app-outlet').innerHTML = `
    <div class="card">
      <h2>🎓 Личный кабинет студента</h2>
      <input type="text" id="stu-number" placeholder="Номер диплома">
      <input type="text" id="stu-uni" placeholder="Код вуза">
      <button class="btn" id="btn-find">Найти диплом</button>
    </div>
    <div id="stu-qr-area" class="card hidden">
      <h3>Ваш цифровой сертификат</h3>
      <div class="qr-container"><img id="qr-img" src="" alt="QR Code"></div>
      <label>Срок действия (TTL):</label>
      <select id="qr-ttl">
        <option value="3600">1 час</option><option value="86400" selected>24 часа</option><option value="604800">7 дней</option>
      </select>
      <button class="btn" id="btn-gen-qr">Сгенерировать QR</button>
      <button class="btn" id="btn-share" style="background:#10b981">🔗 Поделиться ссылкой</button>
    </div>
  `;

  document.getElementById('btn-find').onclick = async () => {
    try {
      const res = await api.get(`/api/v1/verify?number=${document.getElementById('stu-number').value}&university_code=${document.getElementById('stu-uni').value}`);
      localStorage.setItem('current_diploma_id', res.id || 'temp-id');
      document.getElementById('stu-qr-area').classList.remove('hidden');
      showToast('Диплом найден!', 'success');
    } catch(e) { showToast(e.message, 'error'); }
  };

  document.getElementById('btn-gen-qr').onclick = async () => {
    const dipId = localStorage.getItem('current_diploma_id') || 'temp-id';
    const ttl = parseInt(document.getElementById('qr-ttl').value);
    
    try {
      const res = await api.post('/api/v1/qr/generate', {  // ← Добавлен ведущий /
        diploma_id: dipId,   // ← Исправлено: dipId вместо diplomaId
        ttl: ttl             // ← Исправлено: ttl вместо ttlSeconds
      });
      
      document.getElementById('qr-img').src = res.qr_image_base64;
      document.getElementById('stu-qr-area').classList.remove('hidden');
      showToast('QR-код сгенерирован!', 'success');
    } catch(e) { 
      showToast('Ошибка: ' + e.message, 'error'); 
      console.error('QR generation error:', e);
    }
  };

  document.getElementById('btn-share').onclick = () => {
    navigator.clipboard.writeText(window.location.href).then(() => showToast('Ссылка скопирована!', 'success'));
  };
}