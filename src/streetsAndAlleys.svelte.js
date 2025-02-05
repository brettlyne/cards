import { cardDeck, shuffle, dealStreetsAlleys } from "./utils/utils";

export const streets = $state({
  rows: [],
  foundations: [],
});

export function reset() {
  streets.rows = dealStreetsAlleys(shuffle([...cardDeck]));
  streets.foundations = [];
}

// Initialize the game
reset();
