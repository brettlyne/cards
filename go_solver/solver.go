package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"
	"sync"
	"time"
	"container/heap"
)

// Card represents a playing card
type Card struct {
	Value string // A, 2-10, J, Q, K
	Suit  string // H, D, C, S
}

// GameState represents the current state of the game
type GameState struct {
	Rows       [][]Card
	Foundation [][]Card
}

// CompactState is a string representation of the game state
type CompactState string

// Generate a unique hash for the state
func (s CompactState) Hash() uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// Global variables
var (
	stateHistory sync.Map // map[uint64]*StateHistory
	visited      sync.Map // CompactState -> bool
	bestPath     []string
	bestMutex    sync.Mutex
)

// Card to solver byte conversion
func cardToSolver(card Card) byte {
	valueMap := map[string]int{
		"A": 0, "2": 1, "3": 2, "4": 3, "5": 4, "6": 5, "7": 6,
		"8": 7, "9": 8, "10": 9, "J": 10, "Q": 11, "K": 12,
	}
	suitOffset := map[string]int{
		"H": 0, "D": 13, "C": 26, "S": 39,
	}
	
	value := valueMap[card.Value]
	offset := suitOffset[card.Suit]
	return byte('A' + value + offset) // Start at 'A' (65)
}

// Solver byte to Card conversion
func solverToCard(b byte) Card {
	if b == 0 {
		return Card{}
	}
	
	b = b - 'A' // Convert back from ASCII
	suit := "H"
	if b >= 39 {
		suit = "S"
		b -= 39
	} else if b >= 26 {
		suit = "C"
		b -= 26
	} else if b >= 13 {
		suit = "D"
		b -= 13
	}
	
	value := ""
	switch b {
	case 0:
		value = "A"
	case 10:
		value = "J"
	case 11:
		value = "Q"
	case 12:
		value = "K"
	default:
		value = fmt.Sprintf("%d", b+1)
	}
	
	return Card{Value: value, Suit: suit}
}

// Convert game state to compact representation
func (state GameState) ToCompact() CompactState {
	var rows []string
	for _, row := range state.Rows {
		var rowStr string
		for _, card := range row {
			rowStr += string(cardToSolver(card))
		}
		rows = append(rows, rowStr)
	}
	
	var foundations []string
	for _, foundation := range state.Foundation {
		var foundationStr string
		for _, card := range foundation {
			foundationStr += string(cardToSolver(card))
		}
		if foundationStr != "" {
			foundations = append(foundations, foundationStr)
		}
	}
	
	result := strings.Join(rows, "|")
	if len(foundations) > 0 {
		result += "_" + strings.Join(foundations, "_")
	}
	
	return CompactState(result)
}

// Helper functions for move validation
func getSolverValue(card byte) int {
	return int(card-'A') % 13
}

func getSolverSuit(card byte) int {
	return int(card-'A') / 13
}

func canMoveSolver(card, targetCard byte) bool {
	if targetCard == 0 {
		return true
	}
	return getSolverValue(targetCard) == getSolverValue(card)+1
}

func canMoveToFoundationSolver(card byte, foundation []byte) bool {
	if len(foundation) == 0 {
		return getSolverValue(card) == 0 // Ace
	}
	topCard := foundation[len(foundation)-1]
	return getSolverSuit(card) == getSolverSuit(topCard) &&
		getSolverValue(card) == getSolverValue(topCard)+1
}

// StateHistory tracks the best known path to each state
type StateHistory struct {
	parentState CompactState
	move       string
	depth      int // track depth to identify shorter paths
}

// tryAddState attempts to add a state to history if it's new or found a better path
// Returns true if the state should be processed (new or better path found)
func tryAddState(state, parent CompactState, move string, depth int) bool {
	stateHash := state.Hash()
    
    for {
        // Try to load existing state
        val, exists := stateHistory.Load(stateHash)
        if !exists {
            // State doesn't exist, try to store it
            history := &StateHistory{
                parentState: parent,
                move:       move,
                depth:      depth,
            }
            if stateHistory.CompareAndSwap(stateHash, nil, history) {
                return true
            }
            continue // Race condition, try again
        }
        
        existing := val.(*StateHistory)
        if depth < existing.depth {
            // Found a better path, try to update
            history := &StateHistory{
                parentState: parent,
                move:       move,
                depth:      depth,
            }
            if stateHistory.CompareAndSwap(stateHash, existing, history) {
                return true
            }
            continue // Race condition, try again
        }
        return false // Existing path is better or equal
    }
}

// StateItem represents a game state with its priority/depth
type StateItem struct {
	state    CompactState
	priority int    // Used for both score and depth
}

// PriorityQueue implementation
type PriorityQueue []StateItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Higher priority (more cards in foundations) comes first
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(StateItem)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// Worker state now uses priority queue
type WorkerState struct {
	queue PriorityQueue
	lock  sync.Mutex
}

// Add a state to the worker's queue
func (ws *WorkerState) addToQueue(item StateItem) {
	ws.lock.Lock()
	defer ws.lock.Unlock()
	heap.Push(&ws.queue, item)
}

// Get next state from worker's queue
func (ws *WorkerState) getNextState() (StateItem, bool) {
	ws.lock.Lock()
	defer ws.lock.Unlock()
	
	if ws.queue.Len() == 0 {
		return StateItem{}, false
	}
	
	return heap.Pop(&ws.queue).(StateItem), true
}

// Process next states and add them to the queue
func (ws *WorkerState) processNextState(state CompactState, depth int) {
	nextStates := state.NextStates()
	fmt.Printf("Generated %d next states\n", len(nextStates))
	
	for _, nextState := range nextStates {
		// Compare states to determine move type
		oldParts := strings.Split(string(state), "_")
		newParts := strings.Split(string(nextState), "_")
		
		oldFoundations := ""
		newFoundations := ""
		if len(oldParts) > 1 {
			oldFoundations = oldParts[1]
		}
		if len(newParts) > 1 {
			newFoundations = newParts[1]
		}
		
		// Create descriptive move string
		var move string
		if oldFoundations != newFoundations {
			// Find which foundation changed
			oldFounds := strings.Split(oldFoundations, "_")
			newFounds := strings.Split(newFoundations, "_")
			suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
			for i := 0; i < len(suits); i++ {
				oldF := ""
				newF := ""
				if i < len(oldFounds) {
					oldF = oldFounds[i]
				}
				if i < len(newFounds) {
					newF = newFounds[i]
				}
				if oldF != newF {
					if len(newF) > len(oldF) {
						card := solverToCard(newF[len(newF)-1])
						move = fmt.Sprintf("Move %s%s to %s foundation", card.Value, card.Suit, suits[i])
					}
					break
				}
			}
		} else {
			// Compare rows to find the move
			oldRows := strings.Split(oldParts[0], "|")
			newRows := strings.Split(newParts[0], "|")
			for i := range oldRows {
				if len(oldRows[i]) > len(newRows[i]) {
					// This row lost a card
					card := solverToCard(oldRows[i][len(oldRows[i])-1])
					for j := range newRows {
						if len(newRows[j]) > len(oldRows[j]) {
							// This row gained a card
							move = fmt.Sprintf("Move %s%s from row %d to row %d", card.Value, card.Suit, i, j)
							break
						}
					}
					break
				}
			}
		}
		
		if move == "" {
			move = "Unknown move"
		}
		
		// Try to add state, only process if it's new or we found a better path
		if tryAddState(nextState, state, move, depth+1) {
			score := getStateScore(nextState)
			ws.addToQueue(StateItem{nextState, score})
		}
	}
}

// Try to spawn a new worker if queue is too big
func (ws *WorkerState) trySpawnNewWorker(id int, workerStates []*WorkerState, wg *sync.WaitGroup, ctx context.Context, results chan<- []string) bool {
	if id >= len(workerStates)-1 {
		return false // No more worker slots available
	}
	
	nextWorker := workerStates[id+1]
	nextWorker.lock.Lock()
	defer nextWorker.lock.Unlock()
	
	if nextWorker.queue.Len() == 0 { // Only spawn if worker isn't active
		// Transfer states to new worker
		newQueue := make(PriorityQueue, 0, 10000)
		ws.lock.Lock()
		for i := 0; i < 10000 && ws.queue.Len() > 0; i++ {
			item := heap.Pop(&ws.queue).(StateItem)
			heap.Push(&newQueue, item)
		}
		ws.lock.Unlock()
		
		nextWorker.queue = newQueue
		heap.Init(&nextWorker.queue)
		
		fmt.Printf("\nWorker %d: Queue size %d exceeded threshold, spawning worker %d with %d states\n",
			id, ws.queue.Len()+10000, id+1, 10000)
		
		wg.Add(1)
		go worker(id+1, "", results, workerStates, wg, ctx)
		return true
	}
	return false
}

// reconstructPath builds the solution path using shared state history
func reconstructPath(finalState CompactState) []string {
	var path []string
	current := finalState
	visited := make(map[uint64]bool)
	
	for {
		stateHash := current.Hash()
		if visited[stateHash] {
			break // Cycle detected
		}
		visited[stateHash] = true
		
		val, exists := stateHistory.Load(stateHash)
		if !exists {
			break
		}
		
		history := val.(*StateHistory)
		if history.move != "" {
			path = append([]string{history.move}, path...)
		}
		
		if history.parentState == "" {
			break // Reached initial state
		}
		current = history.parentState
	}
	
	return path
}

// Get next possible states from current state
func (state CompactState) NextStates() []CompactState {
	parts := strings.Split(string(state), "_")
	rows := strings.Split(parts[0], "|")
	var foundations []string
	if len(parts) > 1 {
		foundations = strings.Split(parts[1], "_")
	}
	
	// Ensure we have 4 foundations (might be empty)
	for len(foundations) < 4 {
		foundations = append(foundations, "")
	}
	
	nextStates := make([]CompactState, 0, 20)
	
	// Try moving to foundations
	for fromRow := range rows {
		if len(rows[fromRow]) == 0 {
			continue
		}
		
		card := rows[fromRow][len(rows[fromRow])-1]
		for suit := range foundations {
			foundation := []byte(foundations[suit])
			if canMoveToFoundationSolver(card, foundation) {
				// Create new state with the move
				newRows := make([]string, len(rows))
				copy(newRows, rows)
				newFoundations := make([]string, len(foundations))
				copy(newFoundations, foundations)
				
				// Move card
				newRows[fromRow] = rows[fromRow][:len(rows[fromRow])-1]
				newFoundations[suit] = foundations[suit] + string(card)
				
				// Create new state string
				newState := strings.Join(newRows, "|")
				foundationStr := strings.Join(newFoundations, "_")
				if foundationStr != "" {
					newState += "_" + foundationStr
				}
				
				nextState := CompactState(newState)
				nextStates = append(nextStates, nextState)
			}
		}
	}
	
	// Try moving between rows
	for fromRow := range rows {
		if len(rows[fromRow]) == 0 {
			continue
		}
		
		card := rows[fromRow][len(rows[fromRow])-1]
		for toRow := range rows {
			if fromRow == toRow {
				continue
			}
			
			var targetCard byte
			if len(rows[toRow]) > 0 {
				targetCard = rows[toRow][len(rows[toRow])-1]
			}
			
			if canMoveSolver(card, targetCard) {
				// Create new state with the move
				newRows := make([]string, len(rows))
				copy(newRows, rows)
				
				// Move card
				newRows[fromRow] = rows[fromRow][:len(rows[fromRow])-1]
				newRows[toRow] = rows[toRow] + string(card)
				
				// Create new state string
				newState := strings.Join(newRows, "|")
				if len(parts) > 1 {
					newState += "_" + parts[1]
				}
				
				nextState := CompactState(newState)
				nextStates = append(nextStates, nextState)
			}
		}
	}
	
	return nextStates
}

// Check if state is a solution
func isSolution(state CompactState) bool {
	parts := strings.Split(string(state), "_")
	if len(parts) < 2 {
		return false
	}
	
	// All rows should be empty
	rows := strings.Split(parts[0], "|")
	for _, row := range rows {
		if len(row) > 0 {
			return false
		}
	}
	
	// All cards should be in foundations
	foundations := strings.Split(parts[1], "_")
	cardCount := 0
	for _, foundation := range foundations {
		cardCount += len(foundation)
		
		// Check if foundation is properly ordered
		for i := 1; i < len(foundation); i++ {
			prev := foundation[i-1]
			curr := foundation[i]
			if getSolverSuit(prev) != getSolverSuit(curr) ||
				getSolverValue(curr) != getSolverValue(prev)+1 {
				return false
			}
		}
	}
	
	return cardCount == 52 // All cards should be in foundations
}

// Calculate score based on number of cards in foundations
func getStateScore(state CompactState) int {
	parts := strings.Split(string(state), "_")
	if len(parts) != 2 {
		return 0
	}
	foundations := strings.Split(parts[1], "_")
	score := 0
	for _, f := range foundations {
		score += len(f)
	}
	return score
}

// Worker function to process states
func worker(id int, initial CompactState, results chan<- []string, workerStates []*WorkerState, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	
	myState := workerStates[id]
	statesExplored := 0
	lastReport := time.Now()
	
	// Initialize with initial state if provided
	if initial != "" {
		score := getStateScore(initial)
		myState.addToQueue(StateItem{initial, score})
		tryAddState(initial, initial, "Initial state", 0)
		fmt.Printf("Worker %d: Starting with initial state\n", id)
	} else {
		fmt.Printf("Worker %d: Starting with empty queue\n", id)
	}
	
	for {
		// Check for solution found or context cancelled
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d: Stopping due to solution found\n", id)
			return
		default:
		}
		
		// Get next state to process
		item, ok := myState.getNextState()
		if !ok {
			// Try to steal work from other workers
			stolen := false
			for targetId, targetState := range workerStates {
				if targetId == id {
					continue
				}
				
				targetState.lock.Lock()
				if targetState.queue.Len() > 1 { // Leave at least one state
					item = heap.Pop(&targetState.queue).(StateItem)
					stolen = true
					fmt.Printf("Worker %d: Stole state from worker %d\n", id, targetId)
				}
				targetState.lock.Unlock()
				
				if stolen {
					myState.addToQueue(item)
					break
				}
			}
			
			if !stolen {
				fmt.Printf("Worker %d: No work left in any worker, exiting\n", id)
				return
			}
			continue
		}
		
		current := item.state
		depth := item.priority // Using priority as depth
		
		// Check if this is a solution
		if isSolution(current) {
			fmt.Printf("\nWorker %d: Found solution!\n", id)
			path := reconstructPath(current)
			results <- path
			return
		}
		
		// Process next states
		myState.processNextState(current, depth)
		
		// Periodic progress report
		statesExplored++
		if time.Since(lastReport) > 5*time.Second {
			fmt.Printf("Worker %d: Explored %d states, queue size: %d\n", 
				id, statesExplored, myState.queue.Len())
			lastReport = time.Now()
			
			// Try to spawn new worker if queue is large
			if myState.queue.Len() > 30000 {
				myState.trySpawnNewWorker(id, workerStates, wg, ctx, results)
			}
		}
	}
}

func consolidatePathData(workerStates []*WorkerState) (int, int, int) {
	activeStates := make(map[uint64]bool)
    
    // Mark states in worker queues as active
    for _, ws := range workerStates {
        ws.lock.Lock()
        for _, item := range ws.queue {
            activeStates[item.state.Hash()] = true
        }
        ws.lock.Unlock()
    }
    
    // Count states before cleanup
    beforeCount := 0
    stateHistory.Range(func(key, value interface{}) bool {
        beforeCount++
        return true
    })
    
    // Remove inactive states except initial state
    removedCount := 0
    stateHistory.Range(func(key, value interface{}) bool {
        stateHash := key.(uint64)
        if !activeStates[stateHash] {
            history := value.(*StateHistory)
            // Don't remove states that are part of a solution path
            if !isPartOfSolutionPath(stateHash) {
                stateHistory.Delete(key)
                removedCount++
            }
        }
        return true
    })
    
    return removedCount, beforeCount, len(activeStates)
}

// Check if a state is part of any solution path
func isPartOfSolutionPath(stateHash uint64) bool {
    if val, exists := stateHistory.Load(stateHash); exists {
        history := val.(*StateHistory)
        return history.depth < 10 // Keep states near the start
    }
    return false
}

func main() {
	// Create initial deck
	deck := make([]Card, 52)
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	suits := []string{"H", "D", "C", "S"}
	idx := 0
	for _, suit := range suits {
		for _, value := range values {
			deck[idx] = Card{Value: value, Suit: suit}
			idx++
		}
	}
	
	// Shuffle deck
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	
	// Deal initial state
	initialState := GameState{
		Rows:       make([][]Card, 8),
		Foundation: make([][]Card, 0),
	}
	
	cardIdx := 0
	for i := range initialState.Rows {
		cardsInRow := 7
		if i%2 == 1 {
			cardsInRow = 6
		}
		initialState.Rows[i] = make([]Card, cardsInRow)
		for j := 0; j < cardsInRow; j++ {
			initialState.Rows[i][j] = deck[cardIdx]
			cardIdx++
		}
	}
	
	// Convert to compact state
	compactInitial := initialState.ToCompact()
	fmt.Println("Initial state:", string(compactInitial))
	
	// Print initial game state in a readable format
	printInitialState(compactInitial)
	
	// Create worker states with priority queues
	maxWorkers := 8
	workerStates := make([]*WorkerState, maxWorkers)
	for i := range workerStates {
		workerStates[i] = &WorkerState{
			queue: make(PriorityQueue, 0, 1000),
		}
		heap.Init(&workerStates[i].queue)
	}
	
	results := make(chan []string, maxWorkers)
	var wg sync.WaitGroup
	
	fmt.Println("\nStarting search with dynamic worker spawning...")
	startTime := time.Now()
	
	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Start with just one worker
	wg.Add(1)
	go worker(0, compactInitial, results, workerStates, &wg, ctx)
	
	// Wait for result
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()
	
	select {
	case path := <-results:
		if path != nil {
			fmt.Println("\nSolution found!")
			fmt.Println("Solution path:")
			for i, entry := range path {
				fmt.Printf("%d. %s\n", i+1, entry)
			}
		} else {
			fmt.Println("\nNo solution found")
		}
	case <-done:
		fmt.Println("\nAll workers finished with no solution")
	}
	
	fmt.Printf("\nSearch completed in %v\n", time.Since(startTime))
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}

func printInitialState(state CompactState) {
	fmt.Printf("\nInitial state: %s\n", string(state))
	
	parts := strings.Split(string(state), "_")
	rows := strings.Split(parts[0], "|")
	
	fmt.Println("\nInitial rows:")
	for i, row := range rows {
		fmt.Printf("Row %d:", i)
		for j := 0; j < len(row); j++ {
			card := solverToCard(row[j])
			fmt.Printf(" %s%s", card.Value, card.Suit)
		}
		fmt.Println()
	}
	
	if len(parts) > 1 {
		foundations := strings.Split(parts[1], "_")
		fmt.Println("\nInitial foundations:")
		suits := []string{"H", "D", "C", "S"}
		for i, foundation := range foundations {
			if i < len(suits) {
				fmt.Printf("%s:", suits[i])
				for j := 0; j < len(foundation); j++ {
					card := solverToCard(foundation[j])
					fmt.Printf(" %s%s", card.Value, card.Suit)
				}
				fmt.Println()
			}
		}
	}
	fmt.Println()
}