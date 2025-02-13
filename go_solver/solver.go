package main

import (
	"fmt"
	"math"
	"math/rand"
)

const (
	maxRolloutLength = 300 // Maximum number of moves in a simulation
	explorationConstant = 1.414 // sqrt(2)
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

    for len(currentNode.Children) > 0 {
        currentNode, move = currentNode.selectChild()
        nextState, _ := currentState.applyMove(move)
        currentState = nextState
    }

    // Expansion phase - if node has been visited before, expand it
    if currentNode.Visits > 0 {
        currentNode.expand(currentState)
        if len(currentNode.Children) > 0 {
            currentNode, move = currentNode.selectChild()
            nextState, _ := currentState.applyMove(move)
            currentState = nextState
        }
    }

    // Simulation phase
    reward, _ := runMonteCarloSimulation(currentState)

    // Backpropagation phase
    currentNode.backpropagate(reward)
}

// runMonteCarloSimulation performs a random playout from the given game state
// Returns a reward (0-1) and the sequence of moves played
func runMonteCarloSimulation(gameState StreetsGame) (float64, []Move) {
	// Make a copy of the game state to modify
	currentState := gameState.Clone()
	moveHistory := make([]Move, 0)
	
	// Keep track of seen states
	seenStates := make(map[string]bool)
	seenStates[currentState.Hash()] = true
	
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

		if len(legalMoves) == 0 {
			// No more moves possible, evaluate position
			cardsInRows := currentState.CountCardsInRows()
			cardsInFoundation := 52 - cardsInRows
			return float64(cardsInFoundation) / 52.0, moveHistory
		}
		
		// Choose random move
		move := legalMoves[rand.Intn(len(legalMoves))]

		// Apply move
		newState, _ := currentState.applyMove(move)
		seenStates[newState.Hash()] = true

		// Update current state
		currentState = newState
		moveHistory = append(moveHistory, move)
	}
	
	// Reached move limit, evaluate final position
	cardsInRows := currentState.CountCardsInRows()
	cardsInFoundation := 52 - cardsInRows
	return float64(cardsInFoundation) / 52.0, moveHistory
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	example := `TS 8D 6C 9S 2H 2C 3H
9D TH QC 5C AC 7D 5D
5S QS 4C 3D KS 7C AH
8S KC JS JC 2D 9C QD
5H 7S TD 6S AD 4H
TC KH 6D 4S 6H KD
3S 7H AS 2S 8C 4D
3C 9H JD 8H JH QH`
	
	var game StreetsGame
	if err := game.FromString(example); err != nil {
		fmt.Printf("Error parsing game: %v\n", err)
		return
	}
	
	fmt.Println("Initial state:")
	game.Print()
	fmt.Printf("Cards in rows: %d\n", game.CountCardsInRows())
	fmt.Println("\nSimulations:\n")
	
	// Run a few simulations
	rootNode := NewMCTSNode(game.Hash(), nil)
	for i := 0; i < 2000; i++ {
		runMCTS(game, rootNode)
		reward := rootNode.TotalReward / float64(rootNode.Visits)
		fmt.Printf("Reward: %.2f (%.0f cards in foundation)\n", reward, reward*52)
	}
}
