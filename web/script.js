class TaskManager {
    constructor() {
        this.token = null;
        this.currentUser = null;
        this.currentEditId = null;
        this.searchTimeout = null;
        this.tasks = [];
        this.init();
    }

    init() {
        this.bindElements();
        this.initEventListeners();
        this.checkAuthState();
    }

    checkAuthState() {
        const token = localStorage.getItem('jwt_token');
        if (token) {
            this.token = token;
            this.toggleUI(true);
            this.loadTasks();
        }
    }

    bindElements() {
        this.elements = {
            authSection: document.getElementById('auth-section'),
            tasksSection: document.getElementById('tasks-section'),
            loginForm: document.getElementById('login-form'),
            registerForm: document.getElementById('register-form'),
            tasksContainer: document.getElementById('tasks-container'),
            errorMessage: document.getElementById('error-message'),
            errorMessageReg: document.getElementById('error-message-reg'),
            usernameInput: document.getElementById('username'),
            passwordInput: document.getElementById('password'),
            logoutBtn: document.getElementById('logout-btn'),
            searchInput: document.getElementById('search-input'),
            modal: document.getElementById('task-modal'),
            modalTitle: document.getElementById('modal-title'),
            openModalBtn: document.getElementById('open-modal-btn'),
            closeModalBtn: document.querySelector('.close-btn'),
            addTaskForm: document.getElementById('add-task-form')
        };
    }

    toggleUI(authenticated) {
        this.elements.authSection.classList.toggle('hidden', authenticated);
        this.elements.tasksSection.classList.toggle('hidden', !authenticated);
    }

    openModal() {
        this.elements.modal.style.display = 'flex';
    }

    closeModal() {
        this.elements.modal.style.display = 'none';
        this.elements.addTaskForm.reset();
        this.currentEditId = null;
        this.elements.modalTitle.textContent = 'Добавить задачу';
    }

    async handleLogin(e) {
        e.preventDefault();
        const username = this.elements.usernameInput.value.trim();
        const password = this.elements.passwordInput.value.trim();

        if (!username || !password) {
            this.showError('Заполните все поля');
            return;
        }

        try {
            const response = await fetch('/api/signin', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({ username, password })
            });

            const data = await response.json();
            if (!response.ok) throw new Error(data.error || 'Ошибка входа');
            
            this.token = data.token;
            this.currentUser = username;
            localStorage.setItem('jwt_token', data.token);
            this.toggleUI(true);
            await this.loadTasks();
        } catch (error) {
            this.showError(error.message);
        }
    }

    async handleRegister(e) {
        e.preventDefault();
        const username = document.getElementById('reg-username').value.trim();
        const password = document.getElementById('reg-password').value.trim();

        if (!username || !password) {
            this.showError('Заполните все поля', true);
            return;
        }

        try {
            const response = await fetch('/api/signup', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({ username, password })
            });

            document.getElementById('reg-username').value = '';
            document.getElementById('reg-password').value = '';

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || 'Ошибка регистрации');
            }

            // Автоматический вход после регистрации
            const loginResponse = await fetch('/api/signin', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({ username, password })
            });

            const loginData = await loginResponse.json();
            if (!loginResponse.ok) throw new Error(loginData.error || 'Ошибка входа');
            
            this.token = loginData.token;
            this.currentUser = username;
            localStorage.setItem('jwt_token', loginData.token);
            this.toggleUI(true);
            await this.loadTasks();

        } catch (error) {
            const errorElement = document.getElementById('error-message-reg');
            errorElement.textContent = error.message;
            errorElement.classList.remove('success', 'hidden');
            errorElement.classList.add('error');
        }
    }

    logout() {
        this.token = null;
        this.currentUser = null;
        localStorage.removeItem('jwt_token');
        this.toggleUI(false);
    }

    showError(message, isRegister = false) {
        const element = isRegister ? this.elements.errorMessageReg : this.elements.errorMessage;
        element.textContent = message;
        element.classList.remove('hidden');
        setTimeout(() => element.classList.add('hidden'), 3000);
    }

    initEventListeners() {
        this.elements.loginForm.addEventListener('submit', e => this.handleLogin(e));
        this.elements.registerForm.addEventListener('submit', e => this.handleRegister(e));
        this.elements.logoutBtn.addEventListener('click', () => this.logout());
        this.elements.openModalBtn.addEventListener('click', () => this.openModal());
        this.elements.closeModalBtn.addEventListener('click', () => this.closeModal());
        this.elements.addTaskForm.addEventListener('submit', e => this.handleTaskSubmit(e));
        this.elements.searchInput.addEventListener('input', () => this.handleSearchInput());
        window.addEventListener('click', (e) => {
            if (e.target === this.elements.modal) this.closeModal();
        });
    }

    handleSearchInput() {
        clearTimeout(this.searchTimeout);
        this.searchTimeout = setTimeout(() => {
            this.loadTasks(this.elements.searchInput.value.trim());
        }, 300);
    }

    async loadTasks(searchQuery = '') {
        try {
            if (!this.token) throw new Error('Не авторизован');
            
            const url = searchQuery 
                ? `/api/tasks?search=${encodeURIComponent(searchQuery)}` 
                : '/api/tasks';
            
            const response = await fetch(url, {
                headers: { 
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) throw new Error('Ошибка загрузки задач');
            
            const data = await response.json();
            this.tasks = data.tasks || [];
            this.renderTasks();
        } catch (error) {
            this.showError(error.message);
        }
    }

    renderTasks() {
        this.elements.tasksContainer.innerHTML = this.tasks.length > 0 
            ? this.tasks.map(task => {
                const formattedDate = task.date 
                    ? `${task.date.substr(6, 2)}.${task.date.substr(4, 2)}.${task.date.substr(0, 4)}`
                    : '';
                    
                return `
                    <div class="task" data-id="${task.id}">
                        <div class="task-content">
                            <h3>${task.title}</h3>
                            <p>Дата: ${formattedDate}</p>
                            ${task.comment ? `<p>Комментарий: ${task.comment}</p>` : ''}
                        </div>
                        <div>
                            <button class="edit-btn">✎</button>
                            <button class="done-btn">✓</button>
                            <button class="delete-btn">×</button>
                        </div>
                    </div>
                `;
            }).join('')
            : '<div class="no-tasks">Задачи не найдены</div>';

        this.addTaskButtonListeners();
    }

    addTaskButtonListeners() {
        document.querySelectorAll('.delete-btn').forEach(button => {
            button.addEventListener('click', e => this.handleDeleteTask(e));
        });

        document.querySelectorAll('.done-btn').forEach(button => {
            button.addEventListener('click', e => this.handleDoneTask(e));
        });

        document.querySelectorAll('.edit-btn').forEach(button => {
            button.addEventListener('click', e => this.handleEditTask(e));
        });
    }

    async handleDeleteTask(e) {
        const taskElement = e.target.closest('.task');
        const taskId = taskElement.dataset.id;

        try {
            const response = await fetch(`/api/task?id=${taskId}`, {
                method: 'DELETE',
                headers: { 
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) throw new Error('Ошибка удаления задачи');
            await this.loadTasks(this.elements.searchInput.value.trim());
        } catch (error) {
            this.showError(error.message);
        }
    }

    async handleDoneTask(e) {
        const taskElement = e.target.closest('.task');
        const taskId = taskElement.dataset.id;

        try {
            const response = await fetch(`/api/task/done?id=${taskId}`, {
                method: 'POST',
                headers: { 
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) throw new Error('Ошибка отметки задачи как выполненной');
            await this.loadTasks(this.elements.searchInput.value.trim());
        } catch (error) {
            this.showError(error.message);
        }
    }

    handleEditTask(e) {
        const taskElement = e.target.closest('.task');
        const taskId = taskElement.dataset.id;
        const task = this.tasks.find(t => t.id === taskId);

        if (!task) {
            this.showError('Задача не найдена');
            return;
        }

        const form = this.elements.addTaskForm;
        form.elements.title.value = task.title;
        
        const dateStr = task.date;
        const formattedDate = dateStr ? `${dateStr.substr(0, 4)}-${dateStr.substr(4, 2)}-${dateStr.substr(6, 2)}` : '';
        form.elements.date.value = formattedDate;
        
        form.elements.repeat.value = task.repeat || '';
        form.elements.comment.value = task.comment || '';
        this.currentEditId = taskId;
        this.elements.modalTitle.textContent = 'Редактировать задачу';
        this.openModal();
    }

    async handleTaskSubmit(e) {
        e.preventDefault();
        const form = e.target;
        const formData = new FormData(form);
        
        const dateStr = formData.get('date');
        const formattedDate = dateStr ? dateStr.replace(/-/g, '') : '';
        
        const taskData = {
            title: formData.get('title'),
            date: formattedDate,
            repeat: formData.get('repeat') || '',
            comment: formData.get('comment') || ''
        };

        try {
            let response;

            if (this.currentEditId) {
                taskData.id = this.currentEditId;
                response = await fetch('/api/task', {
                    method: 'PUT',
                    headers: { 
                        'Authorization': `Bearer ${this.token}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(taskData)
                });
            } else {
                response = await fetch('/api/task', {
                    method: 'POST',
                    headers: { 
                        'Authorization': `Bearer ${this.token}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(taskData)
                });
            }

            if (!response.ok) throw new Error('Ошибка сохранения задачи');
            this.closeModal();
            await this.loadTasks(this.elements.searchInput.value.trim());
        } catch (error) {
            this.showError(error.message);
        }
    }
}

function switchAuthMode(isLogin) {
    document.getElementById('login-form').reset();
    document.getElementById('register-form').reset();
    document.getElementById('error-message').classList.add('hidden');
    document.getElementById('error-message-reg').classList.add('hidden');
    document.getElementById('login-form').classList.toggle('hidden', !isLogin);
    document.getElementById('register-form').classList.toggle('hidden', isLogin);
}

const taskManager = new TaskManager();