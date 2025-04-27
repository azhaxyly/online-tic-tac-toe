let board = Array(9).fill('');
let currentPlayer = 'X';
let gameMode = null;
let ws = null;
let mySymbol = '';
let opponentSymbol = '';

document.addEventListener('DOMContentLoaded', () => {
  console.log('TicTacToe loaded');

  fetch('/api/nickname', { credentials: 'include' })
    .then(res => res.json())
    .then(data => {
      document.getElementById('nickname').textContent = `Hello, ${data.nickname}!`;
    });

  fetch('/api/stats', { credentials: 'include' })
    .then(res => res.json())
    .then(data => {
      document.getElementById('stats').textContent = `Online: ${data.online}, Active Games: ${data.active_games}`;
    });

  document.getElementById('quick-game-btn').addEventListener('click', startQuickGame);
  document.getElementById('offline-game-btn').addEventListener('click', startOfflineGame);
  document.getElementById('play-again-btn').addEventListener('click', playAgain);
  document.getElementById('back-to-main-btn').addEventListener('click', backToMain);

  renderBoard();
});

function renderBoard() {
  const boardDiv = document.getElementById('game-board');

  if (boardDiv.children.length === 0) {
    for (let idx = 0; idx < 9; idx++) {
      const cellDiv = document.createElement('div');
      cellDiv.classList.add('cell');
      cellDiv.dataset.index = idx;
      cellDiv.addEventListener('click', () => handleCellClick(idx));
      boardDiv.appendChild(cellDiv);
    }
  }

  board.forEach((cell, idx) => {
    const cellDiv = boardDiv.children[idx];
    cellDiv.innerHTML = '';

    if (cell) {
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
      updateStatus('Searching for opponent...');
    };

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log('Received WS message:', msg);

    switch (msg.type) {
      case 'match_found':
        mySymbol = msg.symbol;
        opponentSymbol = mySymbol === 'X' ? 'O' : 'X';
        updateStatus(`Matched! You are '${mySymbol}'`);
        board = Array(9).fill('');
        renderBoard();
        break;

      case 'move_made':
        board[msg.cell] = msg.by;
        renderBoard();
        updateStatus(msg.by === mySymbol ? "Opponent's turn" : "Your turn");
        break;

      case 'game_over':
        updateStatus(msg.result === 'draw' ? "Draw!" : `${msg.result} wins!`);
        endGame();
        break;

      case 'opponent_left':
        updateStatus("Opponent disconnected!");
        endGame();
        break;

      case 'match_cancelled':
        updateStatus("Match cancelled");
        backToMain();
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
  board = Array(9).fill('');
  currentPlayer = 'X';
  mySymbol = 'X';
  opponentSymbol = 'O';
  updateStatus('Offline Game started. You are X.');
  renderBoard();
}

function handleCellClick(idx) {
  if (board[idx]) return;

  if (gameMode === 'offline') {
    board[idx] = currentPlayer;
    renderBoard();

    if (checkWin(currentPlayer)) {
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
    [0,1,2], [3,4,5], [6,7,8],
    [0,3,6], [1,4,7], [2,5,8],
    [0,4,8], [2,4,6]
  ];
  return winPatterns.some(pattern =>
    pattern.every(idx => board[idx] === symbol)
  );
}

function updateStatus(text) {
  document.getElementById('game-status').textContent = text;
}

function endGame() {
  document.getElementById('restart-menu').classList.remove('hidden');
}

function playAgain() {
  window.location.reload();
}

function backToMain() {
  window.location.href = '/';
}
