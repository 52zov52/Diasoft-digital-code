const API_BASE = window.location.hostname === 'localhost' ? 'http://localhost:8080' : '';

export const api = {
  async get(path) { return fetch(`${API_BASE}${path}`).then(handleRes); },
  async post(path, body) {
    return fetch(`${API_BASE}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${localStorage.getItem('jwt') || ''}` },
      body: JSON.stringify(body)
    }).then(handleRes);
  }
};

async function handleRes(res) {
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || data.details || 'Запрос отклонен сервером');
  return data;
}