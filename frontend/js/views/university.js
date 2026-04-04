import { api } from '../api.js';
import { showToast } from '../router.js';

export async function render() {
  document.getElementById('app-outlet').innerHTML = `
    <div class="card">
      <h2>Кабинет ВУЗа</h2>
      <p style="margin-bottom:10px; color:#64748b; font-size:13px;">Загрузите реестр (CSV: serial,fio,year,specialty)</p>
      <input type="file" id="csv-file" accept=".csv">
      <button class="btn" id="btn-upload">Загрузить и подписать</button>
      <div id="upload-status" class="hidden" style="margin-top:10px; color:var(--primary); font-weight:500;"></div>
    </div>
    <div class="card">
      <h3>Последние загруженные записи</h3>
      <table id="diploma-table">
        <thead><tr><th>Серийный номер</th><th>Специальность</th><th>Год</th></tr></thead>
        <tbody id="table-body"><tr><td colspan="3" style="text-align:center">Загрузка...</td></tr></tbody>
      </table>
    </div>
  `;

  // Загружаем последние записи при старте
  loadRecentDiplomas();

  document.getElementById('btn-upload').onclick = async () => {
    const file = document.getElementById('csv-file').files[0];
    if (!file) {
      showToast('Выберите CSV файл', 'error');
      return;
    }
    
    const statusEl = document.getElementById('upload-status');
    statusEl.classList.remove('hidden');
    statusEl.textContent = 'Парсинг и отправка...';

    try {
      const text = await file.text();
      const records = parseCSV(text);
      
      if (records.length === 0) {
        throw new Error('CSV файл пуст или имеет неверный формат');
      }

      const formattedRecords = records.map(row => ({
        serial: row.serial?.trim(),
        fio: row.fio?.trim(),
        year: parseInt(row.year, 10),
        specialty: row.specialty?.trim()
      }));

      const universityId = localStorage.getItem('uni_id') || 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';

      await api.post('/api/v1/diplomas/bulk', {
        university_id: universityId,
        records: formattedRecords
      });
      
      updateTable(formattedRecords.slice(0, 5));
      statusEl.textContent = `Успешно: ${formattedRecords.length} записей`;
      showToast('Реестр обновлен', 'success');
      
      // Перезагружаем данные через 1 секунду
      setTimeout(loadRecentDiplomas, 1000);
    } catch(e) {
      statusEl.textContent = `❌ Ошибка: ${e.message}`;
      showToast(e.message, 'error');
      console.error('Upload error:', e);
    }
  };
}

// ✅ НОВАЯ ФУНКЦИЯ для загрузки с сервера
async function loadRecentDiplomas() {
  try {
    const universityId = localStorage.getItem('uni_id') || 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';
    // Пока заглушка — в будущем здесь будет реальный запрос
    // const data = await api.get(`/api/v1/diplomas/recent?university_id=${universityId}`);
    // updateTable(data);
    
    // Временно показываем пустую таблицу
    const tbody = document.getElementById('table-body');
    if (tbody) {
      tbody.innerHTML = '<tr><td colspan="3" style="text-align:center; color:#64748b;">Нет загруженных записей. Загрузите CSV файл выше.</td></tr>';
    }
  } catch(e) {
    console.error('Failed to load recent diplomas:', e);
  }
}

function parseCSV(text) {
  const lines = text.trim().split('\n').filter(line => line.trim());
  if (lines.length < 2) return [];
  
  const headers = lines[0].split(',').map(h => h.trim().toLowerCase());
  
  return lines.slice(1).map(line => {
    const values = line.split(',').map(v => v.trim());
    const obj = {};
    
    headers.forEach((key, i) => {
      if (key === 'university_id') return;
      
      if (['serial', 'fio', 'year', 'specialty'].includes(key)) {
        obj[key] = values[i] || '';
      }
    });
    
    return obj;
  }).filter(obj => Object.keys(obj).length > 0);
}

function updateTable(records) {
  const tbody = document.getElementById('table-body');
  if (!tbody) return;
  
  if (records.length === 0) {
    tbody.innerHTML = '<tr><td colspan="3" style="text-align:center">Нет загруженных записей</td></tr>';
    return;
  }
  
  tbody.innerHTML = records.map(r => `
    <tr>
      <td>${r.serial || 'N/A'}</td>
      <td>${r.specialty || 'N/A'}</td>
      <td>${r.year || 'N/A'}</td>
    </tr>
  `).join('');
}