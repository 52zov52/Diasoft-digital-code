import * as hrView from './views/hr.js';
import * as studentView from './views/student.js';
import * as uniView from './views/university.js';

// 📦 Конфигурация вкладок
const TABS = [
  { route: '#student', label: 'Студент', icon: 'fa-solid fa-user-graduate', view: studentView },
  { route: '#university', label: 'ВУЗ', icon: 'fa-solid fa-building-columns', view: uniView },
  { route: '#hr', label: 'HR / Работодатель', icon: 'fa-solid fa-magnifying-glass', view: hrView }
];

const DEFAULT_ROUTE = '#hr';

// 🎨 Инициализация вкладок
function initTabs() {
  const tabsContainer = document.getElementById('role-tabs');
  if (!tabsContainer) {
    console.error('❌ Контейнер .tabs не найден!');
    return;
  }

  tabsContainer.innerHTML = '';
  TABS.forEach(tab => {
    const btn = document.createElement('button');
    btn.className = 'tab-btn';
    btn.dataset.route = tab.route;
    btn.innerHTML = `<i class="${tab.icon}"></i> ${tab.label}`;
    btn.addEventListener('click', (e) => {
      e.preventDefault();
      window.location.hash = tab.route;
    });
    tabsContainer.appendChild(btn);
  });
}

// 🔄 Загрузка маршрута
async function loadRoute() {
  let hash = window.location.hash;
  
  if (!hash || !TABS.find(t => t.route === hash)) {
    hash = DEFAULT_ROUTE;
  }

  const tab = TABS.find(t => t.route === hash);
  if (!tab) return;

  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.classList.toggle('active', btn.dataset.route === hash);
  });

  if (tab.view && typeof tab.view.render === 'function') {
    try {
      await tab.view.render();
    } catch (err) {
      console.error(`❌ Ошибка рендера ${hash}:`, err);
      showToast('Не удалось загрузить раздел', 'error');
    }
  }
}

// 🔔 Уведомления
export function showToast(msg, type = 'success') {
  const container = document.getElementById('toast-container');
  if (!container) return;

  const toast = document.createElement('div');
  toast.className = `toast ${type}`;

  const icons = {
    success: 'fa-solid fa-circle-check',
    error: 'fa-solid fa-circle-xmark',
    warning: 'fa-solid fa-triangle-exclamation',
    info: 'fa-solid fa-circle-info'
  };

  toast.innerHTML = `<i class="${icons[type] || icons.success}"></i><span>${msg}</span>`;
  container.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateX(100%)';
    setTimeout(() => toast.remove(), 300);
  }, 4000);
}

// 🚀 Инициализация
window.addEventListener('hashchange', loadRoute);
window.addEventListener('DOMContentLoaded', () => {
  initTabs();
  loadRoute();
});