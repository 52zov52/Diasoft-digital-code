import * as hrView from './views/hr.js';
import * as studentView from './views/student.js';
import * as uniView from './views/university.js';

// Конфигурация вкладок
const tabsConfig = [
  { id: '#student', label: 'Студент', icon: 'fa-solid fa-user-graduate', view: studentView },
  { id: '#university', label: 'ВУЗ', icon: 'fa-solid fa-building-columns', view: uniView },
  { id: '#hr', label: 'HR / Работодатель', icon: 'fa-solid fa-magnifying-glass', view: hrView },
];

// Создаём кнопки табов
function initTabs() {
  const container = document.getElementById('role-tabs');
  if (!container) {
    console.error('❌ Container #role-tabs not found');
    return;
  }

  container.innerHTML = '';
  
  tabsConfig.forEach(tab => {
    const btn = document.createElement('button');
    btn.className = 'tab-btn';
    btn.dataset.route = tab.id;
    btn.innerHTML = `<i class="${tab.icon}"></i> <span>${tab.label}</span>`;
    
    btn.addEventListener('click', () => {
      window.location.hash = tab.id;
    });
    
    container.appendChild(btn);
  });
}

// Загрузка маршрута
async function loadRoute() {
  let hash = window.location.hash;
  
  if (!hash) {
    hash = '#hr';
    window.location.hash = hash;
    return;
  }

  const currentTab = tabsConfig.find(t => t.id === hash);
  
  if (currentTab) {
    document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
    const activeBtn = document.querySelector(`.tab-btn[data-route="${hash}"]`);
    if (activeBtn) activeBtn.classList.add('active');

    if (currentTab.view && typeof currentTab.view.render === 'function') {
      await currentTab.view.render();
    }
  }
}

// Уведомления
export function showToast(msg, type = 'success') {
  let container = document.getElementById('toast-container');
  
  if (!container) {
    container = document.createElement('div');
    container.id = 'toast-container';
    document.body.appendChild(container);
  }

  const t = document.createElement('div');
  t.className = `toast ${type}`;
  
  const iconClass = type === 'error' ? 'fa-solid fa-circle-xmark' : 'fa-solid fa-circle-check';
  
  t.innerHTML = `<i class="${iconClass}"></i> <span>${msg}</span>`;
  
  container.appendChild(t);
  
  setTimeout(() => {
    t.style.opacity = '0';
    setTimeout(() => t.remove(), 300);
  }, 4000);
}

// Инициализация при загрузке
window.addEventListener('hashchange', loadRoute);
window.addEventListener('load', () => {
  console.log('🚀 Router initialized');
  initTabs();
  loadRoute();
});