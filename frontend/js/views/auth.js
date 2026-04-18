// Auth View - форма регистрации и входа

let currentUser = null;

export async function render() {
  const outlet = document.getElementById('app-outlet');
  if (!outlet) return;

  outlet.innerHTML = `
    <div class="auth-container">
      <div class="auth-card">
        <h2><i class="fa-solid fa-right-to-bracket"></i> Вход в систему</h2>
        
        <form id="login-form" class="auth-form">
          <div class="form-group">
            <label for="login-username">Логин</label>
            <input type="text" id="login-username" name="login" required 
                   placeholder="Введите ваш логин" minlength="3">
          </div>
          
          <div class="form-group">
            <label for="login-password">Пароль</label>
            <input type="password" id="login-password" name="password" required 
                   placeholder="Введите пароль" minlength="6">
          </div>
          
          <button type="submit" class="btn btn-primary">
            <i class="fa-solid fa-sign-in-alt"></i> Войти
          </button>
        </form>
        
        <div class="auth-divider">
          <span>или</span>
        </div>
        
        <button id="show-register" class="btn btn-secondary">
          <i class="fa-solid fa-user-plus"></i> Регистрация
        </button>
      </div>
      
      <!-- Форма регистрации (скрыта по умолчанию) -->
      <div id="register-card" class="auth-card" style="display: none;">
        <h2><i class="fa-solid fa-user-plus"></i> Регистрация</h2>
        
        <form id="register-form" class="auth-form">
          <div class="form-group">
            <label for="reg-login">Логин</label>
            <input type="text" id="reg-login" name="login" required 
                   placeholder="Придумайте логин" minlength="3" maxlength="50">
          </div>
          
          <div class="form-group">
            <label for="reg-password">Пароль</label>
            <input type="password" id="reg-password" name="password" required 
                   placeholder="Придумайте пароль" minlength="6" maxlength="100">
          </div>
          
          <div class="form-group">
            <label for="reg-role">Роль</label>
            <select id="reg-role" name="role" required>
              <option value="">Выберите роль</option>
              <option value="student">Студент</option>
              <option value="university">ВУЗ</option>
              <option value="hr">HR / Работодатель</option>
            </select>
          </div>
          
          <div id="university-code-group" class="form-group" style="display: none;">
            <label for="reg-university-code">Код ВУЗа</label>
            <input type="text" id="reg-university-code" name="university_code" 
                   placeholder="Введите код вашего ВУЗа" minlength="3" maxlength="16">
            <small class="form-hint">Код ВУЗа необходим для загрузки дипломов через CSV</small>
          </div>
          
          <div class="form-actions">
            <button type="submit" class="btn btn-primary">
              <i class="fa-solid fa-user-check"></i> Зарегистрироваться
            </button>
            <button type="button" id="show-login" class="btn btn-secondary">
              <i class="fa-solid fa-arrow-left"></i> Назад ко входу
            </button>
          </div>
        </form>
      </div>
      
      <!-- Информация о пользователе (после входа) -->
      <div id="user-info" class="auth-card" style="display: none;">
        <h2><i class="fa-solid fa-user-circle"></i> Профиль пользователя</h2>
        <div class="user-details">
          <p><strong>Логин:</strong> <span id="user-login"></span></p>
          <p><strong>Роль:</strong> <span id="user-role"></span></p>
          <p id="user-uni-block" style="display: none;"><strong>Код ВУЗа:</strong> <span id="user-uni-code"></span></p>
        </div>
        <button id="logout-btn" class="btn btn-danger">
          <i class="fa-solid fa-sign-out-alt"></i> Выйти
        </button>
      </div>
    </div>
  `;

  initAuthListeners();
  checkAuth();
}

function initAuthListeners() {
  // Переключение между входом и регистрацией
  document.getElementById('show-register')?.addEventListener('click', () => {
    document.getElementById('login-form').closest('.auth-card').style.display = 'none';
    document.getElementById('register-card').style.display = 'block';
  });

  document.getElementById('show-login')?.addEventListener('click', () => {
    document.getElementById('register-card').style.display = 'none';
    document.querySelector('.auth-card:first-child').style.display = 'block';
  });

  // Показать/скрыть поле кода ВУЗа
  document.getElementById('reg-role')?.addEventListener('change', (e) => {
    const uniGroup = document.getElementById('university-code-group');
    const uniInput = document.getElementById('reg-university-code');
    if (e.target.value === 'university') {
      uniGroup.style.display = 'block';
      uniInput.required = true;
    } else {
      uniGroup.style.display = 'none';
      uniInput.required = false;
    }
  });

  // Обработка входа
  document.getElementById('login-form')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const login = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;

    try {
      const response = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ login, password })
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Ошибка входа');
      }

      // Сохраняем токен и данные пользователя
      localStorage.setItem('authToken', data.token);
      localStorage.setItem('user', JSON.stringify(data.user));
      currentUser = data.user;

      showToast('Успешный вход!', 'success');
      renderUserInfo();
    } catch (err) {
      showToast(err.message, 'error');
    }
  });

  // Обработка регистрации
  document.getElementById('register-form')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const login = document.getElementById('reg-login').value;
    const password = document.getElementById('reg-password').value;
    const role = document.getElementById('reg-role').value;
    const universityCode = document.getElementById('reg-university-code').value;

    const payload = { login, password, role };
    if (role === 'university' && universityCode) {
      payload.university_code = universityCode;
    }

    try {
      const response = await fetch('/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Ошибка регистрации');
      }

      showToast('Регистрация успешна! Теперь войдите.', 'success');
      document.getElementById('show-login').click();
    } catch (err) {
      showToast(err.message, 'error');
    }
  });

  // Выход
  document.getElementById('logout-btn')?.addEventListener('click', () => {
    localStorage.removeItem('authToken');
    localStorage.removeItem('user');
    currentUser = null;
    showToast('Вы вышли из системы', 'info');
    location.reload();
  });
}

async function checkAuth() {
  const storedUser = localStorage.getItem('user');
  const token = localStorage.getItem('authToken');

  if (storedUser && token) {
    try {
      currentUser = JSON.parse(storedUser);
      renderUserInfo();
    } catch (e) {
      localStorage.removeItem('user');
      localStorage.removeItem('authToken');
    }
  }
}

function renderUserInfo() {
  if (!currentUser) return;

  document.getElementById('login-form').closest('.auth-card').style.display = 'none';
  document.getElementById('register-card').style.display = 'none';
  document.getElementById('user-info').style.display = 'block';

  document.getElementById('user-login').textContent = currentUser.login;
  
  const roleNames = {
    'student': 'Студент',
    'university': 'ВУЗ',
    'hr': 'HR / Работодатель'
  };
  document.getElementById('user-role').textContent = roleNames[currentUser.role] || currentUser.role;

  const uniBlock = document.getElementById('user-uni-block');
  if (currentUser.role === 'university' && currentUser.university_code) {
    uniBlock.style.display = 'block';
    document.getElementById('user-uni-code').textContent = currentUser.university_code;
  }
}

export function getCurrentUser() {
  return currentUser;
}

export function getToken() {
  return localStorage.getItem('authToken');
}
