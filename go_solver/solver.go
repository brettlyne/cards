type Card struct {
	Value int    // 1 for Ace, 11 for Jack, 12 for Queen, 13 for King
	Suit  string // "H" for Hearts, "D" for Diamonds, "C" for Clubs, "S" for Spades
}

// streets and alleys game state representation
// any cards not in a row are implicitly in the foundation
// 19 possible cards if a king was dealt on top at de	al
type StreetsGame struct {
		Rows [8][19]Card
}
