package main

import (
	"fmt"
	"math/rand"
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

// Print current game state
func (g *StreetsGame) Print() {
	fmt.Println("―――――――――――――――――――――――――――――――――")
	for row := 0; row < 8; row++ {
		fmt.Printf("row %d: ", row)
		for col := 0; col < 19; col++ {
			card := g.Rows[row][col]
			if (card != Card{}) {
				fmt.Printf("%s ", card.String())
			}
		}
		fmt.Println()
	}
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
func (g *StreetsGame) Equals(other *StreetsGame) bool {
	// Both games must be non-nil
	if other == nil {
		return false
	}
	
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

func main() {
	var game StreetsGame
	game.Reset()
	
	fmt.Println("Game state:")
	game.Print()
	
	hash := game.Hash()
	fmt.Printf("\nHash length: %d bytes\n", len(hash))

	var game2 StreetsGame
	game2.FromHash(hash)
	game2.Print()

}