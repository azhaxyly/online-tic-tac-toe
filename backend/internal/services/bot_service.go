package services

import (
	"math/rand"
	"tictactoe/internal/models"
	"time"
)

type BotService struct {
	rand *rand.Rand
}

func NewBotService() *BotService {
	return &BotService{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetBotMove возвращает ход бота в зависимости от сложности
func (b *BotService) GetBotMove(board [9]string, difficulty models.BotDifficulty, botSymbol string) int {
	switch difficulty {
	case models.DifficultyEasy:
		return b.getEasyMove(board)
	case models.DifficultyMedium:
		return b.getMediumMove(board, botSymbol)
	case models.DifficultyHard:
		return b.getHardMove(board, botSymbol)
	default:
		return b.getEasyMove(board)
	}
}

// getEasyMove - случайный ход
func (b *BotService) getEasyMove(board [9]string) int {
	available := b.getAvailableCells(board)
	if len(available) == 0 {
		return -1
	}
	return available[b.rand.Intn(len(available))]
}

// getMediumMove - 50% optimal, 50% random
func (b *BotService) getMediumMove(board [9]string, botSymbol string) int {
	// 50% шанс сделать умный ход
	if b.rand.Float32() < 0.5 {
		return b.getHardMove(board, botSymbol)
	}
	return b.getEasyMove(board)
}

// getHardMove - minimax алгоритм (непобедимый)
func (b *BotService) getHardMove(board [9]string, botSymbol string) int {
	playerSymbol := "X"
	if botSymbol == "X" {
		playerSymbol = "O"
	}

	bestScore := -1000
	bestMove := -1

	for i := 0; i < 9; i++ {
		if board[i] == "" {
			// Пробуем ход
			testBoard := board
			testBoard[i] = botSymbol
			score := b.minimax(testBoard, 0, false, botSymbol, playerSymbol)

			if score > bestScore {
				bestScore = score
				bestMove = i
			}
		}
	}

	// Если minimax не нашел ход (не должно случиться), делаем случайный
	if bestMove == -1 {
		return b.getEasyMove(board)
	}

	return bestMove
}

// minimax - рекурсивный алгоритм для поиска оптимального хода
func (b *BotService) minimax(board [9]string, depth int, isMaximizing bool, botSymbol, playerSymbol string) int {
	// Проверяем терминальные состояния
	winner, _ := b.checkWinner(board)

	if winner == botSymbol {
		return 10 - depth // Победа бота (приоритет быстрым победам)
	}
	if winner == playerSymbol {
		return depth - 10 // Победа игрока
	}
	if b.isBoardFull(board) {
		return 0 // Ничья
	}

	if isMaximizing {
		// Ход бота - максимизируем счет
		bestScore := -1000
		for i := 0; i < 9; i++ {
			if board[i] == "" {
				testBoard := board
				testBoard[i] = botSymbol
				score := b.minimax(testBoard, depth+1, false, botSymbol, playerSymbol)
				bestScore = max(score, bestScore)
			}
		}
		return bestScore
	} else {
		// Ход игрока - минимизируем счет
		bestScore := 1000
		for i := 0; i < 9; i++ {
			if board[i] == "" {
				testBoard := board
				testBoard[i] = playerSymbol
				score := b.minimax(testBoard, depth+1, true, botSymbol, playerSymbol)
				bestScore = min(score, bestScore)
			}
		}
		return bestScore
	}
}

// checkWinner проверяет победителя на доске
func (b *BotService) checkWinner(board [9]string) (string, []int) {
	winPatterns := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // горизонтали
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // вертикали
		{0, 4, 8}, {2, 4, 6}, // диагонали
	}

	for _, pattern := range winPatterns {
		a, b, c := pattern[0], pattern[1], pattern[2]
		if board[a] != "" && board[a] == board[b] && board[b] == board[c] {
			return board[a], pattern
		}
	}

	return "", nil
}

// isBoardFull проверяет, заполнена ли доска
func (b *BotService) isBoardFull(board [9]string) bool {
	for _, cell := range board {
		if cell == "" {
			return false
		}
	}
	return true
}

// getAvailableCells возвращает список свободных клеток
func (b *BotService) getAvailableCells(board [9]string) []int {
	available := []int{}
	for i, cell := range board {
		if cell == "" {
			available = append(available, i)
		}
	}
	return available
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
