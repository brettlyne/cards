import { solvedGames } from "./solved-games.js";
import readline from "readline";

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
});

const waitForInput = () =>
  new Promise((resolve) => {
    rl.question("Press Enter to continue...", () => {
      resolve();
    });
  });

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

const getLegalMoves = (rows, foundations) => {
  const legalMoves = [];
  for (let i = 0; i < 8; i++) {
    if (rows[i].length === 0) continue;
    for (let j = 0; j < 8; j++) {
      if (i === j) continue;
      if (canMoveToRow(rows[i][rows[i].length - 1], rows[j])) {
        legalMoves.push({ from: i, to: j });
      }
    }
    const card = rows[i][rows[i].length - 1];
    const suit = getCardSuit(card);
    if (foundations[suit] + 1 === getCardValue(card)) {
      legalMoves.push({ from: i, to: -1 });
    }
  }

  return legalMoves;
};

const applyMove = (rows, foundations, move) => {
  const { from, to } = move;
  const returnValue = JSON.parse(JSON.stringify({ rows, foundations }));
  console.log("🚀 ~ file: optimize-paths.js:56 ~ returnValue:", returnValue);
  console.log("🚀 ~ file: optimize-paths.js:58 ~ from:", from);
  const card = returnValue.rows[from].pop();

  if (to === -1) {
    returnValue.foundations[getCardSuit(card)] += 1;
  } else {
    returnValue.rows[to].push(card);
  }
  return returnValue;
};

const optimizePath = async (g) => {
  const { game, moves } = g;
  const rows = game.split("\n").map((row) => row.split(" "));
  console.log("🚀 ~ file: optimize-paths.js:70 ~ rows:", rows);
  const foundations = { S: 0, H: 0, D: 0, C: 0 }; // track top card
  const possibleMoves = getLegalMoves(rows, foundations);
  console.log(
    "🚀 ~ file: optimize-paths.js:58 ~ possibleMoves:",
    possibleMoves
  );
  for (const move of possibleMoves) {
    const newState = applyMove(rows, foundations, move);
    console.log(
      "🚀 ~ file: optimize-paths.js:63 ~ optimizePath ~ newState:",
      newState
    );
  }

  let tree = {}; // key is state, val is parent state
  tree[JSON.stringify({ rows, foundations })] = null;
  for (const move of possibleMoves) {
    const newState = applyMove(rows, foundations, move);
    tree[JSON.stringify(newState)] = JSON.stringify({ rows, foundations });
  }

  let currState = { rows, foundations };
  for (const move of moves) {
    await waitForInput();

    console.log("🚀 ~ file: optimize-paths.js:96 ~ move:", {
      from: move[0],
      to: move[1],
    });
    const nextState = applyMove(currState.rows, currState.foundations, {
      from: move[0],
      to: move[1],
    });
    const key = JSON.stringify(nextState);
    const possibleMoves = getLegalMoves(nextState.rows, nextState.foundations);
    for (const possibleMove of possibleMoves) {
      const possibleState = applyMove(
        nextState.rows,
        nextState.foundations,
        possibleMove
      );
      const possibleKey = JSON.stringify(possibleState);
      if (!tree[possibleKey]) {
        tree[possibleKey] = key;
      }
    }
    currState = nextState;
  }

  console.log(
    tree[
      JSON.stringify({
        rows: [[], [], [], [], [], [], [], [], [], []],
        foundations: { S: 13, H: 13, D: 13, C: 13 },
      })
    ]
  );
};

const main = async () => {
  await optimizePath(solvedGames[0]);
  rl.close();
};

main().catch(console.error);
