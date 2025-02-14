package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Card struct {
	Value int    // 1 for Ace, 11 for Jack, 12 for Queen, 13 for King
	Suit  string // "H" for Hearts, "D" for Diamonds, "C" for Clubs, "S" for Spades
}

// Convert card to string representation (e.g., "AS" for Ace of Spades)
func (c Card) String() string {
	if (c == Card{}) { // Empty card
		return "  "
	}
	
	value := ""
	switch c.Value {
	case 1:
		value = "A"
	case 10:
		value = "T"
	case 11:
		value = "J"
	case 12:
		value = "Q"
	case 13:
		value = "K"
	default:
		value = fmt.Sprintf("%d", c.Value)
	}
	
	return value + c.Suit
}

// streets and alleys game state representation
// any cards not in a row are implicitly in the foundation
type StreetsGame struct {
	Rows [8][19]Card
}

// Creates a new deck of 52 cards
func createDeck() []Card {
	suits := []string{"H", "D", "C", "S"}
	deck := make([]Card, 52)
	i := 0
	
	for _, suit := range suits {
		for value := 1; value <= 13; value++ {
			deck[i] = Card{Value: value, Suit: suit}
			i++
		}
	}
	
	return deck
}

// Shuffles a deck of cards using Fisher-Yates algorithm
func shuffleDeck(deck []Card) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for i := len(deck) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
}

// Reset deals a new shuffled deck into the game layout
// 4 rows of 7 cards and 4 rows of 6 cards
func (g *StreetsGame) Reset() {
	// Clear current game state
	g.Rows = [8][19]Card{}
	
	// Create and shuffle deck
	deck := createDeck()
	shuffleDeck(deck)
	
	// Deal cards
	cardIndex := 0
	for row := 0; row < 8; row++ {
		cardsInRow := 7
		if row >= 4 {
			cardsInRow = 6
		}
		
		for col := 0; col < cardsInRow; col++ {
			g.Rows[row][col] = deck[cardIndex]
			cardIndex++
		}
	}
}

// ToString converts the game state to a string representation
func (g *StreetsGame) ToString() string {
	var result strings.Builder
	
	for row := 0; row < 8; row++ {
		if row > 0 {
			result.WriteString("\n")
		}
		
		firstCard := true
		for col := 0; col < 19; col++ {
			card := g.Rows[row][col]
			if (card != Card{}) {
				if !firstCard {
					result.WriteString(" ")
				}
				result.WriteString(card.String())
				firstCard = false
			}
		}
	}
	
	return result.String()
}

// FromString reconstructs a game state from its string representation
func (g *StreetsGame) FromString(s string) error {
	// Clear current state
	g.Rows = [8][19]Card{}
	
	// Split into rows, handling both Unix and Windows line endings
	rows := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if len(rows) > 8 {
		return fmt.Errorf("too many rows: %d", len(rows))
	}
	
	for rowNum, row := range rows {
		if len(strings.TrimSpace(row)) == 0 {
			continue
		}
		
		cards := strings.Fields(row)
		if len(cards) > 19 {
			return fmt.Errorf("too many cards in row %d: %d", rowNum, len(cards))
		}
		
		for colNum, cardStr := range cards {
			if len(cardStr) != 2 {
				return fmt.Errorf("invalid card format at row %d, col %d: %s", rowNum, colNum, cardStr)
			}
			
			// Parse value
			var value int
			switch cardStr[0] {
			case 'A':
				value = 1
			case 'T':
				value = 10
			case 'J':
				value = 11
			case 'Q':
				value = 12
			case 'K':
				value = 13
			default:
				if cardStr[0] < '2' || cardStr[0] > '9' {
					return fmt.Errorf("invalid card value at row %d, col %d: %s", rowNum, colNum, cardStr)
				}
				value = int(cardStr[0] - '0')
			}
			
			// Parse suit
			suit := string(cardStr[1])
			if suit != "H" && suit != "D" && suit != "C" && suit != "S" {
				return fmt.Errorf("invalid suit at row %d, col %d: %s", rowNum, colNum, cardStr)
			}
			
			g.Rows[rowNum][colNum] = Card{Value: value, Suit: suit}
		}
	}
	
	return nil
}

// Print displays the current game state
func (g *StreetsGame) Print() {
	fmt.Println(g.ToString())
}

// Get row length (number of non-empty cards)
func (g *StreetsGame) getRowLength(row int) int {
	length := 0
	for col := 0; col < 19; col++ {
		if (g.Rows[row][col] != Card{}) {
			length++
		}
	}
	return length
}

// Get first card in row, or empty card if row is empty
func (g *StreetsGame) getFirstCard(row int) Card {
	for col := 0; col < 19; col++ {
		if (g.Rows[row][col] != Card{}) {
			return g.Rows[row][col]
		}
	}
	return Card{}
}

// Compare two cards for sorting (lower value first, then H, D, C, S)
func compareCards(a, b Card) bool {
	if a.Value != b.Value {
		return a.Value < b.Value
	}
	
	// Map suits to priority (H=0, D=1, C=2, S=3)
	suitPriority := map[string]int{"H": 0, "D": 1, "C": 2, "S": 3}
	return suitPriority[a.Suit] < suitPriority[b.Suit]
}

// Normalize rows: longest first, break ties with first card (lowest value, then H,D,C,S)
func (g *StreetsGame) NormalizeRows() {
	// Create index array to track original positions
	indices := make([]int, 8)
	for i := range indices {
		indices[i] = i
	}
	
	// Sort indices based on row criteria
	for i := 0; i < 7; i++ {
		for j := i + 1; j < 8; j++ {
			lenI := g.getRowLength(indices[i])
			lenJ := g.getRowLength(indices[j])
			
			// If lengths are different, longer row comes first
			if lenI < lenJ {
				indices[i], indices[j] = indices[j], indices[i]
				continue
			}
			
			// If lengths are equal, compare first cards
			if lenI == lenJ {
				cardI := g.getFirstCard(indices[i])
				cardJ := g.getFirstCard(indices[j])
				
				// If either row is empty, push it to the end
				if (cardI == Card{}) {
					indices[i], indices[j] = indices[j], indices[i]
					continue
				}
				if (cardJ == Card{}) {
					continue
				}
				
				// Compare cards
				if !compareCards(cardI, cardJ) {
					indices[i], indices[j] = indices[j], indices[i]
				}
			}
		}
	}
	
	// Create new array with sorted rows
	newRows := [8][19]Card{}
	for newPos, oldPos := range indices {
		copy(newRows[newPos][:], g.Rows[oldPos][:])
	}
	
	// Update game state
	g.Rows = newRows
}

// Hash generates a compact string representation of the normalized game state
// Each card is encoded as a 6-bit number (0-51) where:
// - Value is (card.Value - 1) * 4 (0-48)
// - Suit adds 0-3 (Hearts=0, Diamonds=1, Clubs=2, Spades=3)
// Rows are separated by the delimiter value 63 (111111 in binary)
func (g *StreetsGame) Hash() string {
	// First normalize the state
	g.NormalizeRows()
	
	// Pre-calculate suit values for faster lookup
	suitValue := map[string]byte{"H": 0, "D": 1, "C": 2, "S": 3}
	
	// Each card and delimiter takes 6 bits
	result := make([]byte, 0, 40) // Max capacity needed
	
	var accumulator uint32
	var bitsInAccumulator uint8
	
	// Helper to add 6 bits to our bit stream
	add6Bits := func(value byte) {
		accumulator = (accumulator << 6) | uint32(value)
		bitsInAccumulator += 6
		
		// While we have 8 or more bits, extract bytes
		for bitsInAccumulator >= 8 {
			result = append(result, byte(accumulator>>(bitsInAccumulator-8)))
			bitsInAccumulator -= 8
			accumulator &= (1 << bitsInAccumulator) - 1
		}
	}
	
	// Process all cards
	for row := 0; row < 8; row++ {
		// Add cards in this row
		for col := 0; col < 19; col++ {
			card := g.Rows[row][col]
			if (card != Card{}) {
				cardValue := byte((card.Value-1)*4) + suitValue[card.Suit]
				add6Bits(cardValue)
			}
		}
		
		// Add row delimiter (63 = 111111 in binary)
		if row < 7 { // Don't need delimiter after last row
			add6Bits(63)
		}
	}
	
	// Add any remaining bits with padding
	if bitsInAccumulator > 0 {
		accumulator <<= (8 - bitsInAccumulator)
		result = append(result, byte(accumulator))
	}
	
	return string(result)
}

// FromHash reconstructs a game state from its hash representation
func (g *StreetsGame) FromHash(hash string) error {
	// Clear current state
	g.Rows = [8][19]Card{}
	
	// Convert string back to bytes
	data := []byte(hash)
	if len(data) == 0 {
		return fmt.Errorf("empty hash")
	}
	
	// Reverse lookup for suits
	suitFromValue := []string{"H", "D", "C", "S"}
	
	// Track current position
	var accumulator uint32
	var bitsInAccumulator uint8
	currentRow := 0
	currentCol := 0
	
	// Helper to get next 6 bits
	getNext6Bits := func() (byte, error) {
		// Fill accumulator if needed
		for bitsInAccumulator < 6 && len(data) > 0 {
			accumulator = (accumulator << 8) | uint32(data[0])
			bitsInAccumulator += 8
			data = data[1:]
		}
		
		if bitsInAccumulator < 6 {
			if bitsInAccumulator == 0 {
				return 0, nil // Clean end of data
			}
			return 0, fmt.Errorf("incomplete card data")
		}
		
		// Extract top 6 bits
		result := byte(accumulator >> (bitsInAccumulator - 6)) & 0x3F
		bitsInAccumulator -= 6
		accumulator &= (1 << bitsInAccumulator) - 1
		
		return result, nil
	}
	
	// Process all bytes
	for {
		cardValue, err := getNext6Bits()
		if err != nil {
			return err
		}
		if cardValue == 0 && len(data) == 0 && bitsInAccumulator == 0 {
			break // Clean end of data
		}
		
		if cardValue == 63 {
			// Row delimiter
			currentRow++
			currentCol = 0
			if currentRow >= 8 {
				return fmt.Errorf("too many rows")
			}
			continue
		}
		
		// Convert 6-bit value back to card
		cardNum := int(cardValue / 4) + 1
		suit := suitFromValue[cardValue%4]
		
		if currentCol >= 19 {
			return fmt.Errorf("too many cards in row %d", currentRow)
		}
		
		g.Rows[currentRow][currentCol] = Card{Value: cardNum, Suit: suit}
		currentCol++
	}
	
	return nil
}

// Equals compares this game state with another game state
func (g StreetsGame) Equals(other StreetsGame) bool {
	// First normalize both states
	g.NormalizeRows()
	other.NormalizeRows()
	
	// Compare each position
	for row := 0; row < 8; row++ {
		for col := 0; col < 19; col++ {
			if g.Rows[row][col] != other.Rows[row][col] {
				return false
			}
		}
	}
	
	return true
}

const (
	// Foundation represents moving a card to the foundation
	Foundation = -1
)

// Move represents moving a card from one row to another
// If To is Foundation, the card is moved to the foundation
type Move struct {
	From int
	To   int
}

// String returns a human-readable representation of the move
func (m Move) String() string {
	if m.To == Foundation {
		return fmt.Sprintf("from row %d to foundation", m.From)
	}
	return fmt.Sprintf("from row %d to row %d", m.From, m.To)
}

// getLastCard returns the last card in a row and its column position
// If the row is empty, returns an empty card and -1
func (g *StreetsGame) getLastCard(row int) (Card, int) {
	for col := 18; col >= 0; col-- {
		if (g.Rows[row][col] != Card{}) {
			return g.Rows[row][col], col
		}
	}
	return Card{}, -1
}

// getLowestRemainingCards returns a map of suit to lowest remaining card value
func (g *StreetsGame) getLowestRemainingCards() map[string]int {
	lowest := map[string]int{
		"H": 14, // Higher than any card
		"D": 14,
		"C": 14,
		"S": 14,
	}
	
	// Check all cards in all rows
	for row := 0; row < 8; row++ {
		for col := 0; col < 19; col++ {
			card := g.Rows[row][col]
			if (card != Card{}) {
				if card.Value < lowest[card.Suit] {
					lowest[card.Suit] = card.Value
				}
			}
		}
	}
	
	return lowest
}

// generateLegalMoves returns all legal moves in the current game state
func (g *StreetsGame) generateLegalMoves() []Move {
	moves := make([]Move, 0)
	lowest := g.getLowestRemainingCards()
	
	// Find first empty row if any
	emptyRow := -1
	for row := 0; row < 8; row++ {
		if _, col := g.getLastCard(row); col == -1 {
			emptyRow = row
			break
		}
	}
	
	// For each row, get the last card
	for fromRow := 0; fromRow < 8; fromRow++ {
		card, col := g.getLastCard(fromRow)
		if col == -1 { // Empty row
			continue
		}
		
		// Check if this card is the lowest of its suit
		if card.Value == lowest[card.Suit] {
			moves = append(moves, Move{From: fromRow, To: Foundation})
		}
		
		// If we found an empty row, we can move there
		if emptyRow != -1 && emptyRow != fromRow {
			moves = append(moves, Move{From: fromRow, To: emptyRow})
		}
		
		// Check if this card can move to another non-empty row
		for toRow := 0; toRow < 8; toRow++ {
			if fromRow == toRow {
				continue
			}
			
			targetCard, targetCol := g.getLastCard(toRow)
			if targetCol == -1 { // Skip empty rows
				continue
			}
			
			// Can only move to a card one higher
			if card.Value+1 == targetCard.Value {
				moves = append(moves, Move{From: fromRow, To: toRow})
			}
		}
	}
	
	return moves
}

// Clone returns a deep copy of the game state
func (g StreetsGame) Clone() StreetsGame {
	var clone StreetsGame
	clone.Rows = g.Rows // This works because arrays are copied by value in Go
	return clone
}

// applyMove returns a new game state that results from applying the move
func (g *StreetsGame) applyMove(move Move) (StreetsGame, error) {
	// Create a copy of the game state
	newState := g.Clone()
	
	// Get the card we're moving
	card, fromCol := newState.getLastCard(move.From)
	if fromCol == -1 {
		return newState, fmt.Errorf("invalid move: source row %d is empty", move.From)
	}
	
	// Remove card from source row
	newState.Rows[move.From][fromCol] = Card{}
	
	// If moving to foundation, we're done
	if move.To == Foundation {
		return newState, nil
	}
	
	// Otherwise, add card to destination row
	// Find first empty slot in destination row
	for col := 0; col < 19; col++ {
		if (newState.Rows[move.To][col] == Card{}) {
			newState.Rows[move.To][col] = card
			return newState, nil
		}
	}
	
	return newState, fmt.Errorf("invalid move: destination row %d is full", move.To)
}

// CountCardsInRows returns the total number of cards still in the rows (not in foundations)
func (g *StreetsGame) CountCardsInRows() int {
	count := 0
	for row := 0; row < 8; row++ {
		for col := 0; col < 19; col++ {
			if (g.Rows[row][col] != Card{}) {
				count++
			}
		}
	}
	return count
}

// func main() {
// 	example := `TS 8D 6C 9S 2H 2C 3H
// 9D TH QC 5C AC 7D 5D
// 5S QS 4C 3D KS 7C AH
// 8S KC JS JC 2D 9C QD
// 5H 7S TD 6S AD 4H
// TC KH 6D 4S 6H KD
// 3S 7H AS 2S 8C 4D
// 3C 9H JD 8H JH QH`
	
// 	var game StreetsGame
// 	if err := game.FromString(example); err != nil {
// 		fmt.Printf("Error parsing game: %v\n", err)
// 		return
// 	}

// 	fmt.Println("Initial state from example string:")
// 	game.Print()

// 	fmt.Println("\nState after normalizing:")
// 	game.NormalizeRows()
// 	game.Print()
	
// 	// Test hash roundtrip
// 	hash := game.Hash()
// 	var gameTwo StreetsGame
// 	if err := gameTwo.FromHash(hash); err != nil {
// 		fmt.Printf("Error parsing game in hash test: %v\n", err)
// 		return
// 	}
// 	if !game.Equals(gameTwo) {
// 		fmt.Println("Hash roundtrip failed")
// 	} else {
// 		fmt.Println("\nHash Test roundtrip succeeded")
// 	}
	
// 	// Get and print legal moves
// 	fmt.Println("\nLegal moves:")
// 	moves := game.generateLegalMoves()
// 	for i, move := range moves {
// 		if move.To == Foundation {
// 			fromCard, _ := game.getLastCard(move.From)
// 			fmt.Printf("%d: %s (value: %d)\n", i, move, fromCard.Value)
// 		} else {
// 			fromCard, _ := game.getLastCard(move.From)
// 			toCard, _ := game.getLastCard(move.To)
// 			fmt.Printf("%d: %s (%d -> %d)\n", i, move, fromCard.Value, toCard.Value)
// 		}
// 	}
	
// 	// Apply a foundation move and a regular move
// 	for i, move := range moves {
// 		fmt.Printf("\nApplying move %d: %s\n", i, move)
// 		newState, err := game.applyMove(move)
// 		if err != nil {
// 			fmt.Printf("Error applying move: %v\n", err)
// 			continue
// 		}
// 		newState.Print()
		
// 		// Just do two moves for demonstration
// 		if i == 1 {
// 			break
// 		}
// 	}
// }