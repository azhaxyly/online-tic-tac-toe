const BACKEND_HOST = 'online-tic-tac-toe-qsah.onrender.com';
const API_URL = `https://${BACKEND_HOST}`;
const WS_URL = `wss://${BACKEND_HOST}`;
let board = Array(9).fill('');
let currentPlayer = 'X';
let gameMode = null;
let ws = null;
let mySymbol = '';
let opponentSymbol = '';
let wins = 0;
let losses = 0;
let draws = 0;
let hasRematched = false;
let rematchTimerId = null;
let currentTurn = '';
let moveTimerInterval = null;
let moveDeadline = null;
const MOVE_TIMEOUT = 15;
const REMATCH_DURATION = 15;

let statsInterval = null;
let opponentNickname = '';
let myNickname = '';
let activeSkin = 'default';


// DOM элементы
let authContainer, loginForm, registerForm, showRegisterLink, showLoginLink, authErrorLogin, authErrorRegister;
let userInfoDiv, nicknameSpan, logoutBtn, statsDiv, menuDiv;

// --- Инициализация приложения ---
document.addEventListener('DOMContentLoaded', () => {
  // Кэшируем DOM элементы
  authContainer = document.getElementById('auth-container');
  loginForm = document.getElementById('login-form');
  registerForm = document.getElementById('register-form');
  showRegisterLink = document.getElementById('show-register-link');
  showLoginLink = document.getElementById('show-login-link');
  authErrorLogin = document.getElementById('auth-error-login');
  authErrorRegister = document.getElementById('auth-error-register');

  userInfoDiv = document.getElementById('user-info');
  nicknameSpan = document.getElementById('nickname');
  logoutBtn = document.getElementById('logout-btn');
  statsDiv = document.getElementById('stats');
  menuDiv = document.getElementById('menu');

  // Обработчики форм
  loginForm.addEventListener('submit', handleLogin);
  registerForm.addEventListener('submit', handleRegister);
  logoutBtn.addEventListener('click', handleLogout);

  showRegisterLink.addEventListener('click', (e) => {
    e.preventDefault();
    loginForm.classList.add('hidden');
    registerForm.classList.remove('hidden');
    setAuthError('', 'login');
  });

  showLoginLink.addEventListener('click', (e) => {
    e.preventDefault();
    registerForm.classList.add('hidden');
    loginForm.classList.remove('hidden');
    setAuthError('', 'register');
  });

  // Обработчики кнопок игры
  document.getElementById('quick-game-btn').addEventListener('click', startQuickGame);
  document.getElementById('offline-game-btn').addEventListener('click', startOfflineGame);
  document.getElementById('bot-game-btn').addEventListener('click', showBotMenu);
  document.getElementById('back-from-bot-btn').addEventListener('click', hideBotMenu);
  document.getElementById('play-again-btn').addEventListener('click', playAgain);
  document.getElementById('back-to-main-btn').addEventListener('click', backToMain);
  document.getElementById('cancel-search-btn').addEventListener('click', cancelSearch);
  document.getElementById('leaderboard-btn').addEventListener('click', showLeaderboard);
  document.getElementById('close-leaderboard').addEventListener('click', hideLeaderboard);
  document.getElementById('leaderboard-modal').addEventListener('click', (e) => {
    if (e.target === document.getElementById('leaderboard-modal')) {
      hideLeaderboard();
    }
  });

  // Shop Listeners
  document.getElementById('shop-btn').addEventListener('click', loadShop);
  document.getElementById('close-shop').addEventListener('click', () => {
    document.getElementById('shop-modal').classList.add('hidden');
  });
  document.getElementById('shop-modal').addEventListener('click', (e) => {
    if (e.target === document.getElementById('shop-modal')) {
      document.getElementById('shop-modal').classList.add('hidden');
    }
  });
  document.getElementById('watch-ad-btn').addEventListener('click', watchAd);

  document.querySelectorAll('.difficulty-card').forEach(card => {
    card.addEventListener('click', () => {
      const difficulty = card.dataset.difficulty;
      startBotGame(difficulty);
    });
  });

  checkLoginStatus();
});

// --- Функции аутентификации ---

async function checkLoginStatus() {
  try {
    const res = await fetch(`${API_URL}/api/nickname`, { credentials: 'include' });
    if (!res.ok) {
      throw new Error('Not logged in');
    }
    const data = await res.json();
    showAuthenticatedUI(data.nickname);
    myNickname = data.nickname;

    // Восстановление прерванной онлайн-игры
    const saved = JSON.parse(localStorage.getItem('savedGame'));
    if (saved && saved.gameMode === 'online') {
      console.log('Restoring saved online game...');
      restoreOnlineGame(saved);
    }
  } catch (err) {
    showLoggedOutUI();
  }
}

async function handleLogin(e) {
  e.preventDefault();
  setAuthError('', 'login');

  const nickname = document.getElementById('login-nickname').value.trim();
  const password = document.getElementById('login-password').value;

  if (!nickname || !password) {
    setAuthError('Please fill in all fields', 'login');
    return;
  }

  try {
    const res = await fetch(`${API_URL}/api/login`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ nickname, password }),
    });

    if (!res.ok) {
      const errData = await res.json();
      throw new Error(errData.error || 'Login failed');
    }

    const data = await res.json();
    showAuthenticatedUI(data.nickname);
    myNickname = data.nickname;
  } catch (err) {
    setAuthError(err.message, 'login');
  }
}

async function handleRegister(e) {
  e.preventDefault();
  setAuthError('', 'register');

  const nickname = document.getElementById('register-nickname').value.trim();
  const password = document.getElementById('register-password').value;

  if (!nickname || !password) {
    setAuthError('Please fill in all fields', 'register');
    return;
  }

  if (nickname.length < 3) {
    setAuthError('Nickname must be at least 3 characters', 'register');
    return;
  }

  if (password.length < 6) {
    setAuthError('Password must be at least 6 characters', 'register');
    return;
  }

  try {
    const res = await fetch(`${API_URL}/api/register`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ nickname, password }),
    });

    if (!res.ok) {
      const contentType = res.headers.get("content-type");
      if (contentType && contentType.includes("application/json")) {
        const errData = await res.json();
        throw new Error(errData.error || 'Registration failed');
      } else {
        throw new Error(`Server error: ${res.statusText}`);
      }
    }

    setAuthError('Registration successful! Please log in.', 'register', 'success');
    registerForm.classList.add('hidden');
    loginForm.classList.remove('hidden');
    document.getElementById('login-nickname').value = nickname;
    document.getElementById('login-password').focus();
  } catch (err) {
    console.error("Registration failed:", err);
    if (err.message.includes("JSON")) {
      setAuthError("Failed to communicate with server. Check your connection.", 'register');
    } else {
      setAuthError(err.message, 'register');
    }
  }
}

async function handleLogout() {
  try {
    await fetch(`${API_URL}/api/logout`, { method: 'POST', credentials: 'include' });
  } catch (err) {
    console.error('Logout error:', err);
  }
  window.location.reload();
}

function setAuthError(message, formType, type = 'error') {
  const el = (formType === 'login') ? authErrorLogin : authErrorRegister;
  el.textContent = message;
  el.className = `auth-error ${type}`;
}

function showAuthenticatedUI(nickname) {
  authContainer.classList.add('hidden');
  userInfoDiv.classList.remove('hidden');
  nicknameSpan.textContent = `Hello, ${nickname}!`;
  statsDiv.classList.remove('hidden');
  menuDiv.classList.remove('hidden');

  if (statsInterval) clearInterval(statsInterval);
  loadStats();
  statsInterval = setInterval(loadStats, 60000);
}

function showLoggedOutUI() {
  userInfoDiv.classList.add('hidden');
  statsDiv.classList.add('hidden');
  menuDiv.classList.add('hidden');
  document.getElementById('game-board').classList.add('hidden');
  document.getElementById('restart-menu').classList.add('hidden');

  authContainer.classList.remove('hidden');
  loginForm.classList.remove('hidden');
  registerForm.classList.add('hidden');

  localStorage.removeItem('savedGame');
  if (statsInterval) clearInterval(statsInterval);
}

// --- Логика игры ---

async function loadStats() {
  try {
    const res = await fetch(`${API_URL}/api/stats`, { credentials: 'include' });
    if (!res.ok) {
      if (res.status === 401 || res.status === 403) {
        window.location.reload();
      }
      throw new Error('Failed to fetch stats');
    }
    const data = await res.json();

    let statsText = '';
    if (data.online > 0) {
      statsText += `Online: ${data.online}`;
    }
    if (data.active_games > 0) {
      if (statsText.length > 0) statsText += ' | ';
      statsText += `Active Games: ${data.active_games}`;
    }
    statsDiv.textContent = statsText || 'No active users or games';
  } catch (err) {
    console.error('Failed to load stats:', err);
  }
}

function renderBoard() {
  const boardDiv = document.getElementById('game-board');

  // Apply active skin
  boardDiv.className = 'board'; // reset
  if (activeSkin !== 'default') {
    boardDiv.classList.add(activeSkin);
    document.body.className = activeSkin; // Optional: apply to body for full theme
  } else {
    document.body.className = '';
  }
  if (document.getElementById('game-board').classList.contains('hidden')) {
    boardDiv.classList.add('hidden');
  }

  if (boardDiv.querySelectorAll('.cell').length === 0) {
    for (let idx = 0; idx < 9; idx++) {
      const cellDiv = document.createElement('div');
      cellDiv.classList.add('cell');
      cellDiv.dataset.index = idx;
      cellDiv.addEventListener('click', () => handleCellClick(idx));
      cellDiv.addEventListener('mouseenter', () => {
        if (!board[idx] && !cellDiv.querySelector('.preview')) {
          if (gameMode === 'offline' || (gameMode === 'online' && mySymbol === getCurrentTurn())) {
            const previewSpan = document.createElement('span');
            previewSpan.classList.add('preview', getCurrentTurn() === 'X' ? 'x' : 'o');
            cellDiv.appendChild(previewSpan);
          }
        }
      });
      cellDiv.addEventListener('mouseleave', () => {
        const preview = cellDiv.querySelector('.preview');
        if (preview) preview.remove();
      });
      boardDiv.appendChild(cellDiv);
    }

    // Добавляем линии сетки
    const vLine1 = document.createElement('div');
    vLine1.classList.add('vertical-line');
    const vLine2 = document.createElement('div');
    vLine2.classList.add('vertical-line-right');
    const hLine1 = document.createElement('div');
    hLine1.classList.add('horizontal-line');
    const hLine2 = document.createElement('div');
    hLine2.classList.add('horizontal-line-bottom');
    boardDiv.appendChild(vLine1);
    boardDiv.appendChild(vLine2);
    boardDiv.appendChild(hLine1);
    boardDiv.appendChild(hLine2);
  }

  const cells = boardDiv.querySelectorAll('.cell');
  board.forEach((cell, idx) => {
    const cellDiv = cells[idx];
    const preview = cellDiv.querySelector('.preview');
    if (preview) preview.remove();

    if (cell && !cellDiv.querySelector('span:not(.preview)')) {
      const markSpan = document.createElement('span');
      markSpan.classList.add(cell === 'X' ? 'x' : 'o');
      cellDiv.appendChild(markSpan);
      setTimeout(() => cellDiv.classList.add('show'), 10);
    }
  });
}

async function startQuickGame() {
  gameMode = 'online';

  board = Array(9).fill('');
  const boardDiv = document.getElementById('game-board');
  const cells = boardDiv.querySelectorAll('.cell');
  cells.forEach(cell => {
    cell.innerHTML = '';
    cell.classList.remove('show', 'highlight', 'dim');
  });

  document.querySelector('header').classList.add('hidden');
  userInfoDiv.classList.add('hidden');
  statsDiv.classList.add('hidden');
  menuDiv.classList.add('hidden');
  document.getElementById('cancel-search-btn').classList.remove('hidden');
  hideSideGifs();
  updateStatus('Searching for opponent...');

  ws = new WebSocket(`${WS_URL}/ws`);
  ws.onopen = () => {
    ws.send(JSON.stringify({ type: 'find_match' }));
  };
  setupWebSocketHandlers();
}

function startOfflineGame() {
  gameMode = 'offline';
  board = Array(9).fill('');
  currentPlayer = 'X';
  mySymbol = 'X';
  opponentSymbol = 'O';

  const boardDiv = document.getElementById('game-board');
  const cells = boardDiv.querySelectorAll('.cell');
  cells.forEach(cell => {
    cell.innerHTML = '';
    cell.classList.remove('show', 'highlight', 'dim');
  });
  document.querySelector('header').classList.add('hidden');
  userInfoDiv.classList.add('hidden');
  statsDiv.classList.add('hidden');
  menuDiv.classList.add('hidden');
  document.getElementById('game-board').classList.remove('hidden');
  hideSideGifs();

  showStartScreen();
}

function showBotMenu() {
  menuDiv.classList.add('hidden');
  document.getElementById('bot-menu').classList.remove('hidden');
}

function hideBotMenu() {
  document.getElementById('bot-menu').classList.add('hidden');
  menuDiv.classList.remove('hidden');
}

function startBotGame(difficulty) {
  gameMode = 'online';

  localStorage.removeItem('savedGame');

  // clear board
  board = Array(9).fill('');
  const boardDiv = document.getElementById('game-board');
  const cells = boardDiv.querySelectorAll('.cell');
  cells.forEach(cell => {
    cell.innerHTML = '';
    cell.classList.remove('show', 'highlight', 'dim');
  });

  document.querySelector('header').classList.add('hidden');
  userInfoDiv.classList.add('hidden');
  statsDiv.classList.add('hidden');
  document.getElementById('bot-menu').classList.add('hidden');
  document.getElementById('cancel-search-btn').classList.add('hidden');
  hideSideGifs();

  updateStatus(`Starting game vs Bot (${difficulty})...`);

  ws = new WebSocket(`${WS_URL}/ws`);
  ws.onopen = () => {
    ws.send(JSON.stringify({
      type: 'find_bot_match',
      difficulty: difficulty
    }));
  };
  setupWebSocketHandlers();
}

function handleCellClick(idx) {
  if (board[idx] || !document.getElementById('restart-menu').classList.contains('hidden')) return;

  if (gameMode === 'offline') {
    board[idx] = currentPlayer;
    renderBoard();
    const winningPattern = checkWin(currentPlayer);
    if (winningPattern) {
      highlightWinningCells(winningPattern);
      updateStatus(`${currentPlayer} wins!`);
      endGame();
    } else if (board.every(cell => cell)) {
      updateStatus('Draw!');
      endGame();
    } else {
      currentPlayer = currentPlayer === 'X' ? 'O' : 'X';
      updateStatus(`Turn: ${currentPlayer}`);
    }
  }

  if (gameMode === 'online') {
    if (!ws || ws.readyState !== WebSocket.OPEN || mySymbol !== currentTurn) return;
    board[idx] = mySymbol;
    renderBoard();
    localStorage.setItem('savedGame', JSON.stringify({ gameMode, mySymbol, opponentSymbol, board }));
    ws.send(JSON.stringify({ type: 'move', cell: idx }));
  }
}

function getCurrentTurn() {
  const xCount = board.filter(c => c === 'X').length;
  const oCount = board.filter(c => c === 'O').length;
  return xCount <= oCount ? 'X' : 'O';
}

function checkWin(symbol) {
  const winPatterns = [
    [0, 1, 2], [3, 4, 5], [6, 7, 8],
    [0, 3, 6], [1, 4, 7], [2, 5, 8],
    [0, 4, 8], [2, 4, 6]
  ];
  for (const pattern of winPatterns) {
    if (pattern.every(idx => board[idx] === symbol)) {
      return pattern;
    }
  }
  return null;
}

function highlightWinningCells(winningPattern) {
  const cells = document.querySelectorAll('.cell');
  winningPattern.forEach(idx => {
    cells[idx].classList.add('highlight');
  });
  cells.forEach((cell, idx) => {
    if (!winningPattern.includes(idx)) {
      cell.classList.add('dim');
    }
  });
  setTimeout(() => {
    cells.forEach(cell => {
      cell.classList.remove('highlight', 'dim');
    });
  }, 3000);
}

function updateStatus(text) {
  document.getElementById('game-status').textContent = text;
}

function endGame() {
  document.getElementById('restart-menu').classList.remove('hidden');
}

function playAgain() {
  if (gameMode === 'offline') {
    board = Array(9).fill('');
    currentPlayer = 'X';
    mySymbol = 'X';
    opponentSymbol = 'O';
    const boardDiv = document.getElementById('game-board');
    const cells = boardDiv.querySelectorAll('.cell');
    cells.forEach(cell => {
      cell.innerHTML = '';
      cell.classList.remove('show');
    });
    document.getElementById('restart-menu').classList.add('hidden');
    updateStatus('Offline Game started. You are X.');
    renderBoard();
  } else {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'request_rematch' }));
      updateStatus('Waiting for opponent to accept rematch...');
      document.getElementById('restart-menu').classList.add('hidden');
    }
  }
}

function backToMain() {
  hasRematched = false;
  clearInterval(rematchTimerId);

  localStorage.removeItem('savedGame');

  document.querySelector('header').classList.remove('hidden');
  userInfoDiv.classList.remove('hidden');
  statsDiv.classList.remove('hidden');
  menuDiv.classList.remove('hidden');

  document.getElementById('bot-menu').classList.add('hidden');
  document.getElementById('game-board').classList.add('hidden');
  document.getElementById('restart-menu').classList.add('hidden');
  document.getElementById('cancel-search-btn').classList.add('hidden');

  showSideGifs();

  hidePlayerPanels();

  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.close();
  }

  localStorage.removeItem('savedGame');
}

function cancelSearch() {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: 'cancel_match' }));
    ws.close();
  }
  backToMain();
}

function updateScore() {
  const scoreDiv = document.getElementById('game-score');
  if (!scoreDiv) return;
  scoreDiv.textContent = `Wins: ${wins} | Losses: ${losses} | Draws: ${draws}`;
}

function showRematchDialog() {
  const box = document.getElementById('rematch-box');
  box.classList.remove('hidden');
  const prog = box.querySelector('.timer-progress');
  prog.style.width = '100%';
  prog.classList.remove('blink');

  let remaining = REMATCH_DURATION;
  clearInterval(rematchTimerId);

  rematchTimerId = setInterval(() => {
    remaining--;
    const pct = (remaining / REMATCH_DURATION) * 100;
    prog.style.width = pct + '%';

    if (remaining <= 5) {
      prog.classList.add('blink');
    }

    if (remaining <= 0) {
      clearInterval(rematchTimerId);
      box.classList.add('hidden');
      backToMain();
    }
  }, 1000);

  const acceptBtn = document.getElementById('accept-rematch-btn');
  const declineBtn = document.getElementById('decline-rematch-btn');

  acceptBtn.onclick = () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'accept_rematch' }));
    }
    clearInterval(rematchTimerId);
    box.classList.add('hidden');
  };

  declineBtn.onclick = () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'decline_rematch' }));
    }
    clearInterval(rematchTimerId);
    box.classList.add('hidden');
  };
}

function startNewGame() {
  board = Array(9).fill('');
  const boardDiv = document.getElementById('game-board');
  const cells = boardDiv.querySelectorAll('.cell');
  cells.forEach(cell => {
    cell.innerHTML = '';
    cell.classList.remove('show');
  });
  document.getElementById('restart-menu').classList.add('hidden');
  updateStatus(mySymbol === 'X' ? "Your turn" : "Opponent's turn");
  renderBoard();
}

function restoreOnlineGame(saved) {
  if (!saved || !saved.mySymbol || !saved.opponentSymbol) {
    console.log('Invalid saved game data, clearing...');
    localStorage.removeItem('savedGame');
    return;
  }
  gameMode = 'online';
  mySymbol = saved.mySymbol;
  opponentSymbol = saved.opponentSymbol;
  board = saved.board || Array(9).fill('');

  document.querySelector('header').classList.add('hidden');
  userInfoDiv.classList.add('hidden');
  statsDiv.classList.add('hidden');
  menuDiv.classList.add('hidden');
  document.getElementById('cancel-search-btn').classList.add('hidden');
  document.getElementById('restart-menu').classList.add('hidden');
  document.getElementById('game-board').classList.remove('hidden');

  ws = new WebSocket(`${WS_URL}/ws`);
  ws.onopen = () => {
    console.log('WebSocket reconnected');
    ws.send(JSON.stringify({ type: 'rejoin_match' }));
  };
  setupWebSocketHandlers();

  const stored = parseInt(localStorage.getItem('moveDeadline'), 10);
  if (stored) {
    const now = Date.now();
    if (stored <= now) {
      handleMoveTimeout();
    }
  }
}

function showStartScreen() {
  document.getElementById('game-status').textContent = '';
  const screen = document.getElementById('game-start-screen');
  const text = document.getElementById('game-start-text');
  const board = document.getElementById('game-board');

  screen.classList.remove('hidden');
  text.classList.remove('hidden');
  board.classList.add('hidden');

  setTimeout(() => screen.classList.add('show'), 100);
  setTimeout(() => text.classList.add('show'), 2000);
  setTimeout(() => text.textContent = "GO!", 3200);
  setTimeout(() => {
    screen.classList.add('fade-out');
    text.classList.remove('show');
  }, 3800);
  setTimeout(() => {
    screen.classList.remove('show', 'fade-out');
    screen.classList.add('hidden');
    text.classList.add('hidden');
    board.classList.remove('hidden');
    renderBoard();
    if (gameMode === 'offline') {
      updateStatus('Offline Game started. You are X.');
    }
  }, 4300);
}

function setupWebSocketHandlers() {
  ws.onerror = (err) => {
    console.error('WebSocket error:', err);
    updateStatus('Connection error. Please refresh.');
  };

  ws.onclose = () => {
    console.log('WebSocket closed');
    stopMoveTimer();

    const saved = localStorage.getItem('savedGame');
    if (saved) {
      localStorage.removeItem('savedGame');
      showLoggedOutUI();
      setAuthError("Session expired. Please log in again.", 'login');
    } else if (gameMode === 'online') {
      updateStatus('Disconnected from server');
    }
  };

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log('Received WS message:', msg);

    switch (msg.type) {
      case 'game_state': {
        board = msg.board;
        renderBoard();
        currentTurn = msg.turn;

        if (msg.isFinished) {
          updateStatus(msg.winner ? `${msg.winner} wins!` : 'Draw!');
          endGame();
          localStorage.removeItem('savedGame');
          stopMoveTimer();
        } else {
          updateStatus(currentTurn === mySymbol ? "Your turn" : "Opponent's turn");

          if (currentTurn === mySymbol) {
            const stored = parseInt(localStorage.getItem('moveDeadline'), 10);
            const now = Date.now();
            if (stored && stored > now) {
              startMoveTimer((stored - now) / 1000);
            } else if (stored && stored <= now) {
              handleMoveTimeout();
            } else {
              startMoveTimer();
            }
          } else {
            stopMoveTimer();
          }
        }
        break;
      }

      case 'match_found': {
        mySymbol = msg.symbol;
        opponentNickname = msg.opponent || '';
        opponentSymbol = mySymbol === 'X' ? 'O' : 'X';
        currentTurn = 'X';
        document.getElementById('cancel-search-btn').classList.add('hidden');
        showStartScreen();

        setTimeout(() => {
          document.getElementById('game-board').classList.remove('hidden');
          board = Array(9).fill('');
          renderBoard();
          showPlayerPanels();
          updatePlayerPanels();
          updateStatus(`Matched! You are '${mySymbol}'`);
          localStorage.setItem('savedGame', JSON.stringify({
            gameMode: 'online',
            mySymbol,
            opponentSymbol,
            board
          }));

          if (currentTurn === mySymbol) {
            startMoveTimer();
          }
        }, 4300);
        break;
      }

      case 'move_made': {
        board[msg.cell] = msg.by;
        renderBoard();
        currentTurn = msg.by === 'X' ? 'O' : 'X';
        localStorage.setItem('savedGame', JSON.stringify({
          gameMode,
          mySymbol,
          opponentSymbol,
          board
        }));
        updateStatus(currentTurn === mySymbol ? "Your turn" : "Opponent's turn");

        if (currentTurn === mySymbol) {
          const stored = parseInt(localStorage.getItem('moveDeadline'), 10);
          const now = Date.now();
          if (stored && stored > now) {
            startMoveTimer((stored - now) / 1000);
          } else if (stored && stored <= now) {
            handleMoveTimeout();
          } else {
            startMoveTimer();
          }
        } else {
          stopMoveTimer();
        }
        break;
      }

      case 'game_over': {
        updateStatus(msg.result === 'draw' ? "Draw!" : `${msg.result === mySymbol ? 'You win' : 'You lose'}!`);

        if (msg.result === 'draw') {
          draws++;
        } else if (msg.result === mySymbol) {
          wins++;
        } else {
          losses++;
        }

        updateScore();
        endGame();

        if (msg.result !== 'draw' && msg.winningPattern) {
          setTimeout(() => highlightWinningCells(msg.winningPattern), 200);
        }

        localStorage.removeItem('savedGame');
        stopMoveTimer();
        break;
      }

      case 'opponent_left': {
        updateStatus("Opponent disconnected!");
        endGame();
        stopMoveTimer();
        break;
      }

      case 'rematch_requested':
        showRematchDialog();
        break;

      case 'rematch':
        hasRematched = true;
        mySymbol = msg.symbol;
        opponentSymbol = msg.opponent;
        currentTurn = 'X';
        startNewGame();

        if (currentTurn === mySymbol) {
          startMoveTimer();
        }
        break;

      case 'rematch_declined':
        if (!hasRematched) {
          updateStatus('Opponent declined rematch.');
          setTimeout(backToMain, 3000);
        }
        break;

      case 'error':
        if (msg.message === 'no active game') {
          localStorage.removeItem('savedGame');
          backToMain();
        } else {
          console.error('Server error:', msg.message);
        }
        break;

      default:
        console.warn('Unhandled message type:', msg.type);
    }
  };
}

function startMoveTimer(seconds = MOVE_TIMEOUT) {
  const bar = document.getElementById('move-timer');
  const prog = document.getElementById('move-timer-progress');
  const now = Date.now();

  moveDeadline = now + seconds * 1000;
  localStorage.setItem('moveDeadline', moveDeadline.toString());

  bar.classList.remove('hidden');
  prog.classList.remove('blink-slow', 'blink-med', 'blink-fast');

  clearInterval(moveTimerInterval);

  moveTimerInterval = setInterval(() => {
    const remainingMs = moveDeadline - Date.now();

    if (remainingMs <= 0) {
      clearInterval(moveTimerInterval);
      bar.classList.add('hidden');
      prog.classList.remove('blink-slow', 'blink-med', 'blink-fast');
      localStorage.removeItem('moveDeadline');
      handleMoveTimeout();
    } else {
      const pct = (remainingMs / (MOVE_TIMEOUT * 1000)) * 100;
      prog.style.width = `${pct}%`;
      prog.classList.remove('blink-slow', 'blink-med', 'blink-fast');

      if (remainingMs <= 1000) {
        prog.classList.add('blink-fast');
      } else if (remainingMs <= 3000) {
        prog.classList.add('blink-med');
      } else if (remainingMs <= 5000) {
        prog.classList.add('blink-slow');
      }
    }
  }, 100);
}

function stopMoveTimer() {
  clearInterval(moveTimerInterval);
  document.getElementById('move-timer').classList.add('hidden');
  localStorage.removeItem('moveDeadline');
}

function handleMoveTimeout() {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: 'forfeit' }));
  }
  updateStatus('Time up! You lose.');
  endGame();
  localStorage.removeItem('savedGame');
}

function hideSideGifs() {
  document.querySelectorAll('.side-gif').forEach(img => img.classList.add('hidden'));
}

function showSideGifs() {
  document.querySelectorAll('.side-gif').forEach(img => img.classList.remove('hidden'));
}

// --- Leaderboard Logic ---

async function showLeaderboard() {
  const modal = document.getElementById('leaderboard-modal');
  const tableBody = document.querySelector('#leaderboard-table tbody');
  const loadingText = document.getElementById('leaderboard-loading');

  modal.classList.remove('hidden');
  tableBody.innerHTML = ''; // Очистить старые данные
  loadingText.classList.remove('hidden');

  try {
    const res = await fetch(`${API_URL}/api/leaderboard`, { credentials: 'include' });
    if (!res.ok) throw new Error('Failed to load');

    const data = await res.json();
    loadingText.classList.add('hidden');
    renderLeaderboard(data);
  } catch (err) {
    console.error(err);
    loadingText.textContent = 'Error loading leaderboard';
  }
}

function hideLeaderboard() {
  document.getElementById('leaderboard-modal').classList.add('hidden');
}

function renderLeaderboard(users) {
  const tableBody = document.querySelector('#leaderboard-table tbody');

  if (!users || users.length === 0) {
    tableBody.innerHTML = '<tr><td colspan="5" style="text-align:center; padding: 20px; color: #666">No records yet</td></tr>';
    return;
  }

  users.forEach((user, index) => {
    const tr = document.createElement('tr');

    const rank = index + 1;
    let rankClass = '';
    if (rank === 1) rankClass = 'rank-1';
    if (rank === 2) rankClass = 'rank-2';
    if (rank === 3) rankClass = 'rank-3';

    const totalGames = user.wins + user.losses + user.draws;
    const winrate = totalGames > 0 ? Math.round((user.wins / totalGames) * 100) : 0;
    const eloRating = user.elo_rating || 1000;
    tr.innerHTML = `
      <td class="${rankClass}">${rank}</td>
      <td class="${rankClass}">${escapeHtml(user.nickname)}</td>
      <td class="elo-cell">${eloRating}</td>
  <td>${user.wins}</td>
  <td class="winrate-cell">${winrate}%</td>
`;
    tableBody.appendChild(tr);
  });
}

function escapeHtml(text) {
  if (!text) return text;
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

// --- Shop Logic ---

async function loadShop() {
  const modal = document.getElementById('shop-modal');
  const balanceSpan = document.getElementById('shop-balance-amount');
  const catalogDiv = document.getElementById('shop-catalog');

  modal.classList.remove('hidden');
  catalogDiv.innerHTML = '<p class="loading-text">Loading shop...</p>';

  try {
    const res = await fetch(`${API_URL}/api/shop`, { credentials: 'include' });
    if (!res.ok) throw new Error('Failed to load shop');

    const data = await res.json();
    balanceSpan.textContent = data.coins;
    activeSkin = data.active_skin; // Update global active skin
    renderShopItems(data.catalog, data.inventory, data.active_skin);

    // Update board immediately if skin changed
    renderBoard();
  } catch (err) {
    console.error(err);
    catalogDiv.innerHTML = '<p class="error-text">Error loading shop</p>';
  }
}

function renderShopItems(catalog, inventory, currentSkin) {
  const catalogDiv = document.getElementById('shop-catalog');
  catalogDiv.innerHTML = '';

  // Always add Default skin
  const defaultItem = { id: 'default', name: 'Default', cost: 0, type: 'skin' };
  // Prepend default to catalog for display purposes if not present
  // (Assuming catalog from server doesn't include default, or we handle it manually)

  // Render catalog items
  // We'll treat "default" as a special case if we want it in the list, 
  // but let's stick to the server catalog. 
  // If server catalog doesn't have default, we can add it manually to the UI list.

  const allItems = [
    { id: 'default', name: 'Classic', cost: 0, description: 'Standard look' },
    ...catalog
  ];

  allItems.forEach(item => {
    const itemDiv = document.createElement('div');
    itemDiv.className = 'shop-item';

    const isOwned = item.id === 'default' || (inventory && inventory.includes(item.id));
    const isEquipped = currentSkin === item.id;

    let btnHtml = '';
    if (isEquipped) {
      btnHtml = `<button class="shop-action-btn equipped" disabled>Equipped</button>`;
    } else if (isOwned) {
      btnHtml = `<button class="shop-action-btn equip" onclick="equipItem('${item.id}')">Equip</button>`;
    } else {
      btnHtml = `<button class="shop-action-btn buy" onclick="buyItem('${item.id}')">Buy (${item.cost})</button>`;
    }

    itemDiv.innerHTML = `
      <div class="item-preview ${item.id}">
        <span>X</span>
      </div>
      <span class="item-name">${item.name}</span>
      ${!isOwned ? `<span class="item-cost">🪙 ${item.cost}</span>` : ''}
      ${btnHtml}
    `;
    catalogDiv.appendChild(itemDiv);
  });
}

async function buyItem(itemId) {
  if (!confirm('Buy this item?')) return;

  try {
    const res = await fetch(`${API_URL}/api/shop/buy`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ item_id: itemId })
    });

    if (!res.ok) {
      const err = await res.json();
      alert(err.error || 'Purchase failed');
      return;
    }

    // Refresh shop
    loadShop();
  } catch (err) {
    console.error(err);
    alert('Purchase failed');
  }
}

async function equipItem(itemId) {
  try {
    const res = await fetch(`${API_URL}/api/shop/equip`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ item_id: itemId })
    });

    if (!res.ok) {
      const err = await res.json();
      alert(err.error || 'Equip failed');
      return;
    }

    // Refresh shop and board
    loadShop();
  } catch (err) {
    console.error(err);
    alert('Equip failed');
  }
}

async function watchAd() {
  const btn = document.getElementById('watch-ad-btn');
  const originalText = btn.innerHTML;

  btn.disabled = true;
  btn.innerHTML = 'Watching Ad... (5s)';

  try {
    const res = await fetch(`${API_URL}/api/shop/ad-reward`, {
      method: 'POST',
      credentials: 'include'
    });

    if (!res.ok) throw new Error('Failed to get reward');

    const data = await res.json();
    // alert(`Reward received: +${data.amount} coins!`);

    // Refresh shop balance
    loadShop();
  } catch (err) {
    console.error(err);
    alert('Failed to watch ad');
  } finally {
    btn.disabled = false;
    btn.innerHTML = originalText;
  }
}

// Expose functions to global scope for onclick handlers in HTML string
window.buyItem = buyItem;
window.equipItem = equipItem;