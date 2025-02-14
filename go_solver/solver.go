package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
)

const (
	maxRolloutLength = 150 // Maximum number of moves in a simulation
	// explorationConstant = 1.414 // sqrt(2)
	// √2 is derived from the multi-armed bandit problem
	explorationConstant = 1.8
	iterationsPerMove  = 400 // Number of MCTS iterations to run per move
)

// MCTSNode represents a node in the Monte Carlo Tree Search
type MCTSNode struct {
	GameStateHash string                // Hash of the game state this node represents
	Parent        *MCTSNode             // Pointer to parent node
	Children      map[Move]*MCTSNode    // Map of moves to child nodes
	Visits        int                   // Number of times this node has been visited
	TotalReward   float64               // Sum of rewards from all visits to this node
}

// NewMCTSNode creates a new MCTSNode with initialized fields
func NewMCTSNode(gameStateHash string, parent *MCTSNode) *MCTSNode {
	return &MCTSNode{
		GameStateHash: gameStateHash,
		Parent:        parent,
		Children:      make(map[Move]*MCTSNode),
		Visits:        0,
		TotalReward:   0,
	}
}

// selectChild uses UCT formula to select the most promising child node
func (n *MCTSNode) selectChild() (*MCTSNode, Move) {
    bestScore := -1.0
    var bestChild *MCTSNode
    var bestMove Move

    for move, child := range n.Children {
        if child.Visits == 0 {
            return child, move
        }

        // UCT formula: average reward + exploration bonus
        exploitation := child.TotalReward / float64(child.Visits)
        exploration := explorationConstant * 
            math.Sqrt(math.Log(float64(n.Visits))/float64(child.Visits))
        score := exploitation + exploration

        if score > bestScore {
            bestScore = score
            bestChild = child
            bestMove = move
        }
    }

    return bestChild, bestMove
}

// expand adds all possible child nodes to the current node
func (n *MCTSNode) expand(gameState StreetsGame) {
    legalMoves := gameState.generateLegalMoves()
    
    for _, move := range legalMoves {
        nextState, _ := gameState.applyMove(move)
        if _, exists := n.Children[move]; !exists {
            n.Children[move] = NewMCTSNode(nextState.Hash(), n)
        }
    }
}

// backpropagate updates node statistics up the tree
func (n *MCTSNode) backpropagate(reward float64) {
    current := n
    for current != nil {
        current.Visits++
        current.TotalReward += reward
        current = current.Parent
    }
}

// runMCTS performs one iteration of the MCTS algorithm
func runMCTS(rootState StreetsGame, rootNode *MCTSNode) {
    // Selection phase - traverse tree until we reach a leaf node
    currentNode := rootNode
    currentState := rootState.Clone()
    var move Move
    
    // Track only states in our actual path through the tree
    pathStates := make(map[string]bool)
    pathStates[currentState.Hash()] = true

    // Selection phase
    for len(currentNode.Children) > 0 { // while there are children to visit
        currentNode, move = currentNode.selectChild()
        nextState, _ := currentState.applyMove(move)
        currentState = nextState
        pathStates[currentState.Hash()] = true
    }

    // Expansion phase - if node has been visited before, expand it
    if currentNode.Visits > 0 {
        currentNode.expand(currentState)
        if len(currentNode.Children) > 0 {
            currentNode, move = currentNode.selectChild()
            nextState, _ := currentState.applyMove(move)
            currentState = nextState
            pathStates[currentState.Hash()] = true
        }
    }

    // Simulation phase - now each simulation starts fresh with just the path states
    reward, _ := runMonteCarloSimulation(currentState, pathStates)

    // Backpropagation phase
    currentNode.backpropagate(reward)
}

// runMonteCarloSimulation performs a random playout from the given game state
// Returns a reward (0-1) and the sequence of moves played
func runMonteCarloSimulation(gameState StreetsGame, pathStates map[string]bool) (float64, []Move) {
    // Make a copy of the game state to modify
    currentState := gameState.Clone()
    moveHistory := make([]Move, 0)
    
    // Make a local copy of seen states just for this simulation
    seenStates := make(map[string]bool)
    for state := range pathStates {
        seenStates[state] = true
    }
    
    // Run simulation until we hit max moves or no legal moves remain
    for moveCount := 0; moveCount < maxRolloutLength; moveCount++ {
        // Get legal moves
        legalMoves := currentState.generateLegalMoves()

        // Filter out moves that lead to previously seen states
        validMoves := make([]Move, 0)
        for _, move := range legalMoves {
            nextState, _ := currentState.applyMove(move)
            if !seenStates[nextState.Hash()] {  
                validMoves = append(validMoves, move)
            }
        }

        if len(validMoves) == 0 {
            // All moves lead to previously seen states, evaluate position
            cardsInRows := currentState.CountCardsInRows()
            cardsInFoundation := 52 - cardsInRows
            return float64(cardsInFoundation) / 52.0, moveHistory
        }
        
        // Choose random move from valid moves
        move := validMoves[rand.Intn(len(validMoves))]

        // Apply move
        newState, _ := currentState.applyMove(move)
        seenStates[newState.Hash()] = true  // Only track in local simulation

        // Update current state
        currentState = newState
        moveHistory = append(moveHistory, move)
    }    
    // Reached move limit, evaluate final position
    cardsInRows := currentState.CountCardsInRows()
    cardsInFoundation := 52 - cardsInRows
    return float64(cardsInFoundation + 1) / 52.0, moveHistory //small bonus for reaching move limit
}

// getBestMove returns the move with the highest visit count and its statistics
func (n *MCTSNode) getBestMove() (Move, float64) {
    bestVisits := -1
    var bestMove Move
    bestReward := -1.0
    
    for move, child := range n.Children {
        // If we find a perfect foundation move, return it immediately
        reward := child.TotalReward / float64(child.Visits)
        if reward == 1.0 && move.To == Foundation {
            return move, reward
        }
        if child.Visits > bestVisits {
            bestVisits = child.Visits
            bestMove = move
            bestReward = reward
        }
    }
    return bestMove, bestReward
}

func main() {
    // Read the input file
    content, err := os.ReadFile("winnable_games_fixed.txt")
    if err != nil {
        fmt.Printf("Error reading input file: %v\n", err)
        return
    }
    fmt.Printf("Read %d bytes from input file\n", len(content))

    // Set up logging
    logFile, err := os.OpenFile("winnable_games_moves.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
    if err != nil {
        fmt.Printf("Error opening log file: %v\n", err)
        return
    }
    defer logFile.Close()

    // Split content into games (separated by blank lines)
    // First normalize line endings
    normalizedContent := strings.ReplaceAll(string(content), "\r\n", "\n")
    games := strings.Split(normalizedContent, "\n\n")
    fmt.Printf("Found %d games to analyze\n", len(games))

    for gameNum, gameStr := range games {
        // Skip empty games
        gameStr = strings.TrimSpace(gameStr)
        if gameStr == "" {
            fmt.Printf("Skipping empty game %d\n", gameNum+1)
            continue
        }

        fmt.Printf("\nProcessing game %d (%d lines):\n", gameNum+1, len(strings.Split(gameStr, "\n")))
        fmt.Println(gameStr)

        // Parse the game
        var game StreetsGame
        if err := game.FromString(gameStr); err != nil {
            fmt.Printf("Error parsing game %d: %v\n", gameNum+1, err)
            continue
        }

        // Start solving the game
        currentState := game
        var moves []Move
        
        // Play through the game
        for moveNum := 0; moveNum < 250; moveNum++ {
            rootNode := NewMCTSNode(currentState.Hash(), nil)
            
            // Run MCTS iterations
            for i := 0; i < iterationsPerMove; i++ {
                runMCTS(currentState, rootNode)
            }
            
            // Make the best move
            bestMove, _ := rootNode.getBestMove()
            if bestMove == (Move{}) {
                fmt.Printf("No more moves available after %d moves\n", moveNum)
                break
            }
            
            // Record the move
            moves = append(moves, bestMove)
            
            // Apply the move
            nextState, _ := currentState.applyMove(bestMove)
            currentState = nextState

            // Print progress every 50 moves
            if moveNum % 50 == 0 {
                fmt.Printf("  Made %d moves, cards in rows: %d\n", moveNum, currentState.CountCardsInRows())
            }
        }

        // Log the game and its moves
        result := gameStr + "\nmoves: ["
        for i, move := range moves {
            if i > 0 {
                result += ","
            }
            result += fmt.Sprintf("[%d,%d]", move.From, move.To)
        }
        result += "]\n\n"

        if _, err := logFile.WriteString(result); err != nil {
            fmt.Printf("Error writing to log: %v\n", err)
        }

        fmt.Printf("Completed game %d with %d moves\n", gameNum+1, len(moves))
    }
    
    fmt.Println("Done! Results have been written to winnable_games_moves.log")
}
