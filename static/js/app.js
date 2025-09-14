const API_BASE = '';

let currentUser = null;
let authToken = null;

// Проверяем, есть ли сохраненный токен при загрузке
document.addEventListener('DOMContentLoaded', () => {
    const savedToken = localStorage.getItem('authToken');
    const savedUser = localStorage.getItem('user');
    
    if (savedToken && savedUser) {
        authToken = savedToken;
        currentUser = JSON.parse(savedUser);
        showDashboard();
        loadUserData();
    }
});
async function depositMoney(event) {
    event.preventDefault();
    
    const amount = parseFloat(document.getElementById('deposit-amount').value);
    
    try {
        await apiRequest('/api/deposit', {
            method: 'POST',
            body: JSON.stringify({
                amount: amount
            }),
        });
        
        alert('Счет успешно пополнен!');
        document.getElementById('deposit-amount').value = '';
        
        loadUserData();
    } catch (error) {
        alert('Ошибка пополнения: ' + error.message);
    }
}
// Функции для показа/скрытия форм
function showLogin() {
    document.getElementById('login-form').style.display = 'block';
    document.getElementById('register-form').style.display = 'none';
}

function showRegister() {
    document.getElementById('login-form').style.display = 'none';
    document.getElementById('register-form').style.display = 'block';
}

function showDashboard() {
    document.getElementById('auth-forms').style.display = 'none';
    document.getElementById('dashboard').style.display = 'block';
    document.getElementById('auth-buttons').style.display = 'none';
    document.getElementById('user-info').style.display = 'block';
    document.getElementById('user-name').textContent = currentUser.full_name;
}

function showAuthForms() {
    document.getElementById('auth-forms').style.display = 'block';
    document.getElementById('dashboard').style.display = 'none';
    document.getElementById('auth-buttons').style.display = 'block';
    document.getElementById('user-info').style.display = 'none';
}


// И обновите функцию apiRequest:
async function apiRequest(url, options = {}) {
    const fullUrl = url.startsWith('http') ? url : `http://localhost:8080${url}`;
    
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };
    
    if (authToken) {
        headers['Authorization'] = `Bearer ${authToken}`;
    }
    
    try {
        const response = await fetch(fullUrl, {
            ...options,
            headers,
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || `HTTP error ${response.status}`);
        }
        
        return response.json();
    } catch (error) {
        console.error('API request failed:', error);
        throw error;
    }
}

// Исправляем функции регистрации и входа
async function register(event) {
    event.preventDefault();
    
    const fullName = document.getElementById('register-fullname').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;
    
    try {
        const response = await apiRequest('/auth/register', { // Убрали /api/
            method: 'POST',
            body: JSON.stringify({ 
                email: email, 
                password: password, 
                full_name: fullName 
            }),
        });
        
        authToken = response.token;
        currentUser = response.user;
        
        localStorage.setItem('authToken', authToken);
        localStorage.setItem('user', JSON.stringify(currentUser));
        
        showDashboard();
        loadUserData();
    } catch (error) {
        alert('Ошибка регистрации: ' + error.message);
        console.error('Register error:', error);
    }
}

async function login(event) {
    event.preventDefault();
    
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    
    try {
        const response = await apiRequest('/auth/login', { // Убрали /api/
            method: 'POST',
            body: JSON.stringify({ email: email, password: password }),
        });
        
        authToken = response.token;
        currentUser = response.user;
        
        localStorage.setItem('authToken', authToken);
        localStorage.setItem('user', JSON.stringify(currentUser));
        
        showDashboard();
        loadUserData();
    } catch (error) {
        alert('Ошибка входа: ' + error.message);
        console.error('Login error:', error);
    }
}

function logout() {
    authToken = null;
    currentUser = null;
    localStorage.removeItem('authToken');
    localStorage.removeItem('user');
    showAuthForms();
}

// Функции для работы с данными
async function loadUserData() {
    try {
        const balanceData = await apiRequest('/api/balance');
        document.getElementById('balance-amount').textContent = 
            `${balanceData.balance.toFixed(2)} ${balanceData.currency}`;
        
        const transfers = await apiRequest('/api/transfers');
        renderTransfers(transfers);
    } catch (error) {
        console.error('Ошибка загрузки данных:', error);
    }
}

async function transferMoney(event) {
    event.preventDefault();
    
    const recipientEmail = document.getElementById('recipient-email').value;
    const amount = parseFloat(document.getElementById('transfer-amount').value);
    const currency = document.getElementById('transfer-currency').value;
    
    try {
        await apiRequest('/api/transfer', {
            method: 'POST',
            body: JSON.stringify({
                to_email: recipientEmail,
                amount,
                currency
            }),
        });
        
        alert('Перевод выполнен успешно!');
        document.getElementById('recipient-email').value = '';
        document.getElementById('transfer-amount').value = '';
        
        loadUserData();
    } catch (error) {
        alert('Ошибка перевода: ' + error.message);
    }
}

function renderTransfers(transfers) {
    const container = document.getElementById('transfers-history');
    container.innerHTML = '';
    
    if (transfers.length === 0) {
        container.innerHTML = '<p>Нет операций</p>';
        return;
    }
    
    transfers.forEach(transfer => {
        const div = document.createElement('div');
        div.className = 'transfer-item';
        
        // Определяем направление перевода
        const isOutgoing = transfer.from_account_id === currentUser.id;
        const amountClass = isOutgoing ? 'transfer-negative' : 'transfer-positive';
        const amountPrefix = isOutgoing ? '-' : '+';
        
        div.innerHTML = `
            <div>
                <strong>${isOutgoing ? 'Кому' : 'От'}:</strong> 
                ${isOutgoing ? transfer.to_email : transfer.from_email}
            </div>
            <div class="${amountClass}">
                ${amountPrefix}${transfer.amount} ${transfer.currency}
            </div>
            <div>${new Date(transfer.created_at).toLocaleDateString()}</div>
        `;
        
        container.appendChild(div);
    });
}