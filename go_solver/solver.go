package main

import (
	"fmt"
	"math"
	"math/rand"
)

const (
	maxRolloutLength = 150 // Maximum number of moves in a simulation
	// explorationConstant = 1.414 // sqrt(2)
	// √2 is derived from the multi-armed bandit problem
	explorationConstant = 1.0
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
            // No more valid moves possible, evaluate position
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
    return float64(cardsInFoundation) / 52.0, moveHistory
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
	
	currentState := game
	bestOverallReward := 0.0
	
	// Play through the game making best moves found by MCTS
	for moveNum := 0; moveNum < 200; moveNum++ {
		rootNode := NewMCTSNode(currentState.Hash(), nil)
		
		// Run MCTS iterations
		for i := 0; i < 400; i++ {
			runMCTS(currentState, rootNode)
		}
		
		// Get statistics about all possible moves
		fmt.Printf("\nMove %d - Analysis (Cards in rows: %d):\n", moveNum, currentState.CountCardsInRows())
		for move, child := range rootNode.Children {
			reward := child.TotalReward / float64(child.Visits)
			if reward > bestOverallReward {
				bestOverallReward = reward
			}
			fmt.Printf("  Move %s: visits=%d reward=%.2f\n", move, child.Visits, reward)
		}
		
		// Make the best move
		bestMove, bestReward := rootNode.getBestMove()
		if bestMove == (Move{}) {
			fmt.Println("No moves available")
			break
		}
		
		fmt.Printf("\nChosen move: %s (reward=%.2f)\n", bestMove, bestReward)
		nextState, _ := currentState.applyMove(bestMove)
		currentState = nextState
		
		// Print current state every 10 moves
		if moveNum % 10 == 0 {
			fmt.Println("\nCurrent state:")
			currentState.Print()
		}
	}
	
	fmt.Printf("\nBest reward found in any position: %.2f (%.0f cards in foundation)\n", 
        bestOverallReward, bestOverallReward*52)
}
