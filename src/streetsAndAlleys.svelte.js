import { cardDeck, shuffle } from "./utils/utils";
import { state } from "svelte";

const cardValues = {
  A: 1,
  2: 2,
  3: 3,
  4: 4,
  5: 5,
  6: 6,
  7: 7,
  8: 8,
  9: 9,
  T: 10,
  J: 11,
  Q: 12,
  K: 13,
};

function getCardValue(card) {
  return cardValues[card[0]];
}

function getCardSuit(card) {
  return card[1];
}

function canMoveToRow(card, targetRow) {
  if (targetRow.length === 0) return true;
  const targetCard = targetRow[targetRow.length - 1];
  return getCardValue(targetCard) === getCardValue(card) + 1;
}

function canMoveToFoundation(card, foundation) {
  if (foundation.length === 0) {
    return getCardValue(card) === 1;
  }
  const targetCard = foundation[foundation.length - 1];
  return (
    getCardSuit(card) === getCardSuit(targetCard) &&
    getCardValue(card) === getCardValue(targetCard) + 1
  );
}

// rows alternate between 7 and 6 cards
function dealStreetsAlleys(deck) {
  let result = [];
  for (let i = 0; i < 4; i++) {
    result.push(deck.splice(0, 7));
    result.push(deck.splice(0, 6));
  }
  return result;
}

export const streets = state({
  rows: [],
  foundations: [],
});

export function reset() {
  const deck = shuffle(cardDeck);
  streets.rows = dealStreetsAlleys(deck);
  streets.foundations = [];
}

// Initialize the game
reset();