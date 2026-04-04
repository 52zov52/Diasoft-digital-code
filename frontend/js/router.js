import * as hrView from './views/hr.js';
import * as studentView from './views/student.js';
import * as uniView from './views/university.js';

const routes = { '#student': studentView, '#university': uniView, '#hr': hrView, '': hrView };

window.addEventListener('hashchange', loadRoute);
window.addEventListener('load', loadRoute);

document.querySelectorAll('#role-tabs button').forEach(btn => {
  btn.addEventListener('click', (e) => {
    window.location.hash = e.target.dataset.route;
  });
});

async function loadRoute() {
  const hash = window.location.hash;
  const view = routes[hash] || routes[''];
  document.querySelectorAll('#role-tabs button').forEach(b => b.classList.remove('active'));
  document.querySelector(`#role-tabs button[data-route="${hash}"]`)?.classList.add('active');
  if (view && typeof view.render === 'function') await view.render();
}

export function showToast(msg, type = 'success') {
  const c = document.getElementById('toast-container');
  const t = document.createElement('div');
  t.className = `toast ${type}`;
  t.textContent = msg;
  c.appendChild(t);
  setTimeout(() => t.remove(), 4000);
}