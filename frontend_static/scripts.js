let board = Array(9).fill('');
let currentPlayer = 'X';
let gameMode = null;
let ws = null;
let mySymbol = '';
let opponentSymbol = '';
let winningSymbol = null;
let wins = 0;
let losses = 0;
let draws = 0;

document.addEventListener('DOMContentLoaded', () => {
  console.log('TicTacToe loaded');

  fetch('/api/nickname', { credentials: 'include' })
    .then(res => res.json())
    .then(data => {
      document.getElementById('nickname').textContent = `Hello, ${data.nickname}!`;
    });

  loadStats();
  setInterval(loadStats, 60000);

  document.getElementById('quick-game-btn').addEventListener('click', startQuickGame);
  document.getElementById('offline-game-btn').addEventListener('click', startOfflineGame);
  document.getElementById('play-again-btn').addEventListener('click', playAgain);
  document.getElementById('back-to-main-btn').addEventListener('click', backToMain);
  document.getElementById('cancel-search-btn').addEventListener('click', cancelSearch);

  renderBoard();
});

async function loadStats() {
  try {
    const res = await fetch('/api/stats', { credentials: 'include' });
    if (!res.ok) {
      throw new Error('Failed to fetch stats');
    }
    const data = await res.json();

    const statsDiv = document.getElementById('stats');
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
      setTimeout(() => {
        cellDiv.classList.add('show');
      }, 10);
    }
  });
}

async function startQuickGame() {
  console.log('Starting Quick Game...');
  gameMode = 'online';

  document.querySelector('header').classList.add('hidden');
  document.getElementById('nickname').classList.add('hidden');
  document.getElementById('stats').classList.add('hidden');
  document.getElementById('menu').classList.add('hidden');
  document.getElementById('cancel-search-btn').classList.remove('hidden');

  updateStatus('Searching for opponent...');


  try {
    const res = await fetch('/api/nickname', { credentials: 'include' });
    if (!res.ok) {
      throw new Error('Failed to fetch nickname');
    }
    const data = await res.json();
    console.log('Nickname confirmed:', data.nickname);
  } catch (err) {
    console.error('Cannot start game without nickname/session.', err);
    return;
  }

  ws = new WebSocket(`ws://${location.host}/ws`);

  ws.onopen = () => {
    ws.send(JSON.stringify({ type: 'find_match' }));
  };

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log('Received WS message:', msg);

    switch (msg.type) {
      case 'match_found':
        mySymbol = msg.symbol;
        document.getElementById('cancel-search-btn').classList.add('hidden');
        opponentSymbol = mySymbol === 'X' ? 'O' : 'X';

        showStartScreen();

        setTimeout(() => {
          document.getElementById('game-board').classList.remove('hidden');
          board = Array(9).fill('');
          renderBoard();
          updateStatus(`Matched! You are '${mySymbol}'`);
        }, 4300);
        break;

      case 'move_made':
        board[msg.cell] = msg.by;
        renderBoard();
        updateStatus(msg.by === mySymbol ? "Opponent's turn" : "Your turn");
        break;

      case 'game_over':
        updateStatus(msg.result === 'draw' ? "Draw!" : `${msg.result} wins!`);
        if (msg.result === 'draw') {
          draws++;
        } else if (msg.result === mySymbol) {
          wins++;
        } else {
          losses++;
        }
        updateScore();
        endGame();
        setTimeout(() => {
          if (msg.result !== 'draw' && msg.winningPattern) {
            highlightWinningCells(msg.winningPattern);
          }
        }, 200);
        break;

      case 'opponent_left':
        updateStatus("Opponent disconnected!");
        endGame();
        break;

      case 'match_cancelled':
        updateStatus("Match cancelled");
        backToMain();
        break;
      case 'rematch_requested':
        if (gameMode === 'online') {
          showRematchDialog();
        } else {
          console.warn('Ignored rematch_requested: not in online game.');
        }
        break;

      case 'rematch':
        mySymbol = msg.symbol;
        opponentSymbol = msg.opponent;
        board = Array(9).fill('');
        document.getElementById('restart-menu').classList.add('hidden');
        document.getElementById('game-board').classList.remove('hidden');
        renderBoard();
        updateStatus(mySymbol === 'X' ? "Your turn" : "Opponent's turn");
        break;


      case 'rematch_declined':
        updateStatus('Opponent declined rematch.');
        setTimeout(backToMain, 3000);
        break;

    }
  };

  ws.onclose = () => {
    console.log('WebSocket closed');
    if (gameMode === 'online') {
      updateStatus('Disconnected from server');
    }
  };
}

function startOfflineGame() {
  console.log('Starting Offline Game...');
  gameMode = 'offline';

  document.querySelector('header').classList.add('hidden');
  document.getElementById('nickname').classList.add('hidden');
  document.getElementById('stats').classList.add('hidden');
  document.getElementById('menu').classList.add('hidden');
  document.getElementById('game-board').classList.remove('hidden');

  board = Array(9).fill('');
  currentPlayer = 'X';
  mySymbol = 'X';
  opponentSymbol = 'O';
  showStartScreen();
}

function handleCellClick(idx) {
  if (board[idx]) return;
  if (document.getElementById('restart-menu').classList.contains('hidden') === false) return;

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
    if (!ws || mySymbol !== getCurrentTurn()) return;
    board[idx] = mySymbol;
    renderBoard();
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
  window.location.href = '/';
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
  const modal = document.getElementById('rematch-modal');
  modal.classList.remove('hidden');

  const acceptBtn = document.getElementById('accept-rematch-btn');
  const declineBtn = document.getElementById('decline-rematch-btn');

  acceptBtn.onclick = () => {
    ws.send(JSON.stringify({ type: 'accept_rematch' }));
    modal.classList.add('hidden');
  };

  declineBtn.onclick = () => {
    ws.send(JSON.stringify({ type: 'decline_rematch' }));
    modal.classList.add('hidden');
    backToMain();
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

function showStartScreen() {
  const screen = document.getElementById('game-start-screen');
  const text = document.getElementById('game-start-text');
  const board = document.getElementById('game-board');

  screen.classList.remove('hidden');
  text.classList.remove('hidden');
  board.classList.add('hidden');

  setTimeout(() => {
    screen.classList.add('show');
  }, 100);

  setTimeout(() => {
    text.classList.add('show');
  }, 2000);

  setTimeout(() => {
    text.textContent = "GO!";
  }, 3200);

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
    updateStatus('Offline Game started. You are X.');
  }, 4300);
}