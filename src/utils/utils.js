// prettier-ignore
export const cardDeck = [
  "AC", "2C", "3C", "4C", "5C", "6C", "7C", "8C", "9C", "TC", "JC", "QC", "KC",
  "AD", "2D", "3D", "4D", "5D", "6D", "7D", "8D", "9D", "TD", "JD", "QD", "KD",
  "AS", "2S", "3S", "4S", "5S", "6S", "7S", "8S", "9S", "TS", "JS", "QS", "KS",
  "AH", "2H", "3H", "4H", "5H", "6H", "7H", "8H", "9H", "TH", "JH", "QH", "KH",
];

export const shuffle = (array) => {
  for (let i = array.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [array[i], array[j]] = [array[j], array[i]];
  }
  return array;
};

// our target format is
// - 4x 7-card piles
// - 4x 6-card piles
// - 4x foundation trackers (top card)
// each separate by | character
export const dealStreetsAlleys = (array) => {
  let result = "";
  // 4 piles of 7 cards
  for (let i = 0; i < 4; i++) {
    result += array.slice(0, 7).join("") + "|";
    array = array.slice(7);
  }
  // 4 piles of 6 cards
  for (let i = 0; i < 4; i++) {
    result += array.slice(0, 6).join("") + "|";
    array = array.slice(6);
  }
  // 4 foundation trackers
  result += "0C0D0S0H";
  return result;
};
