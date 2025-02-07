import { cardDeck, shuffle } from "./utils/utils";

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
    return getCardValue(card) === 1; // Ace
  }
  const topCard = foundation[foundation.length - 1];
  return (
    getCardSuit(card) === getCardSuit(topCard) &&
    getCardValue(card) === getCardValue(topCard) + 1
  );
}

function stateToCompact(state) {
  // Convert rows to pipe-separated strings
  const rowsStr = state.rows.map((row) => row.join("")).join("|");

  // For foundations, only need top card of each (or empty string if no cards)
  const foundationsStr = state.foundations
    .map((f) => (f.length ? f[f.length - 1] : ""))
    .join("_");

  return rowsStr + (foundationsStr ? "_" + foundationsStr : "");
}

function compactToState(compact) {
  const [rowsStr, ...foundationsArr] = compact.split("_");

  // Split rows into arrays of 2-char cards
  const rows = rowsStr.split("|").map((row) => {
    const cards = [];
    for (let i = 0; i < row.length; i += 2) {
      cards.push(row.slice(i, i + 2));
    }
    return cards;
  });

  // Convert foundation top cards back to full foundation stacks
  const foundations = foundationsArr.map((topCard) =>
    topCard ? [topCard] : []
  );

  return { rows, foundations };
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

export const streets = $state({
  rows: [],
  foundations: [],
});

export function reset() {
  streets.rows = dealStreetsAlleys(shuffle([...cardDeck]));
  streets.foundations = [];
}

function nextStates(stateStr) {
  const state = compactToState(stateStr);
  const nextStateStrings = [];

  // Try moving each card from each row
  state.rows.forEach((sourceRow, sourceRowIndex) => {
    if (sourceRow.length === 0) return;

    const card = sourceRow[sourceRow.length - 1];

    // Find first empty row if any
    const firstEmptyRowIndex = state.rows.findIndex((row) => row.length === 0);

    // Try moving to other rows
    state.rows.forEach((targetRow, targetRowIndex) => {
      if (sourceRowIndex === targetRowIndex) return;
      if (targetRow.length === 0) {
        // If this is an empty row but not the first empty row, skip it
        if (firstEmptyRowIndex !== -1 && targetRowIndex > firstEmptyRowIndex)
          return;
      }

      if (canMoveToRow(card, targetRow)) {
        const newState = {
          rows: state.rows.map((r, i) => {
            if (i === sourceRowIndex) return r.slice(0, -1);
            if (i === targetRowIndex) return [...r, card];
            return r;
          }),
          foundations: state.foundations,
        };
        nextStateStrings.push(stateToCompact(newState));
      }
    });

    // Try moving to foundations
    state.foundations.forEach((foundation, foundationIndex) => {
      if (canMoveToFoundation(card, foundation)) {
        const newState = {
          rows: state.rows.map((r, i) =>
            i === sourceRowIndex ? r.slice(0, -1) : r
          ),
          foundations: state.foundations.map((f, i) =>
            i === foundationIndex ? [...f, card] : f
          ),
        };
        nextStateStrings.push(stateToCompact(newState));
      }
    });

    // if ace, start a new foundation
    if (getCardValue(card) === 1) {
      const newState = {
        rows: state.rows.map((r, i) =>
          i === sourceRowIndex ? r.slice(0, -1) : r
        ),
        foundations: [...state.foundations, [card]],
      };
      nextStateStrings.push(stateToCompact(newState));
    }
  });

  return nextStateStrings;
}

// Compact card representation for solver only (charCode 48-111)
function cardToSolver(card) {
  const value = getCardValue(card) - 1; // 0-12
  const suit = getCardSuit(card);
  // Leave 16 slots per suit (13 cards + 3 gap)
  if (suit === "H") return String.fromCharCode(48 + value);
  if (suit === "D") return String.fromCharCode(64 + value);
  if (suit === "C") return String.fromCharCode(80 + value);
  if (suit === "S") return String.fromCharCode(96 + value);
}

function solverToCard(char) {
  const code = char.charCodeAt(0);
  if (code < 64) {
    // Hearts (48-63)
    return `${code - 47 === 1 ? "A" : code - 47}H`;
  } else if (code < 80) {
    // Diamonds (64-79)
    return `${code - 63 === 1 ? "A" : code - 63}D`;
  } else if (code < 96) {
    // Clubs (80-95)
    return `${code - 79 === 1 ? "A" : code - 79}C`;
  } else {
    // Spades (96-111)
    return `${code - 95 === 1 ? "A" : code - 95}S`;
  }
}

function getSolverValue(char) {
  return (char.charCodeAt(0) - 48) % 16;
}

function canMoveSolver(card, targetCard) {
  if (!targetCard) return true;
  return getSolverValue(targetCard) === getSolverValue(card) + 1;
}

function getSolverSuit(char) {
  const code = char.charCodeAt(0);
  return code < 64 ? "H" : code < 80 ? "D" : code < 96 ? "C" : "S";
}

function canMoveToFoundationSolver(card, topCard) {
  if (!topCard) return getSolverValue(card) === 0;
  return (
    getSolverSuit(card) === getSolverSuit(topCard) &&
    getSolverValue(card) === getSolverValue(topCard) + 1
  );
}

export function isProbablyWinnable() {
  // Convert initial state to solver format
  function stateToSolver(state) {
    const rowsStr = state.rows
      .map((row) => row.map((card) => cardToSolver(card)).join(""))
      .join("|");

    const foundationsStr = state.foundations
      .map((f) => (f.length ? cardToSolver(f[f.length - 1]) : ""))
      .join("_");

    return rowsStr + (foundationsStr ? "_" + foundationsStr : "");
  }

  function solverToState(str) {
    const [rowsStr, ...foundationsArr] = str.split("_");

    const rows = rowsStr.split("|").map((row) => {
      const cards = [];
      for (let i = 0; i < row.length; i++) {
        cards.push(solverToCard(row[i]));
      }
      return cards;
    });

    const foundations = foundationsArr.map((char) =>
      char ? [solverToCard(char)] : []
    );

    return { rows, foundations };
  }

  function nextSolverStates(solverStr) {
    const [rowsStr, ...foundationsArr] = solverStr.split("_");
    const rows = rowsStr.split("|");
    const nextStates = [];

    // Try moving each card from each row
    rows.forEach((sourceRow, sourceRowIndex) => {
      if (sourceRow.length === 0) return;

      const cardChar = sourceRow[sourceRow.length - 1];
      const firstEmptyRowIndex = rows.findIndex((row) => row.length === 0);

      // Try moving to other rows
      rows.forEach((targetRow, targetRowIndex) => {
        if (sourceRowIndex === targetRowIndex) return;
        if (targetRow.length === 0) {
          if (firstEmptyRowIndex !== -1 && targetRowIndex > firstEmptyRowIndex)
            return;
        }

        const targetChar = targetRow.length
          ? targetRow[targetRow.length - 1]
          : null;
        if (canMoveSolver(cardChar, targetChar)) {
          // Create new rows array with the move
          const newRows = rows.map((r, i) => {
            if (i === sourceRowIndex) return r.slice(0, -1);
            if (i === targetRowIndex) return r + cardChar;
            return r;
          });
          nextStates.push(
            newRows.join("|") +
              (foundationsArr.length ? "_" + foundationsArr.join("_") : "")
          );
        }
      });

      // Try moving to foundations
      foundationsArr.forEach((topChar, foundationIndex) => {
        if (canMoveToFoundationSolver(cardChar, topChar || null)) {
          // Create new state with the move
          const newRows = rows.map((r, i) =>
            i === sourceRowIndex ? r.slice(0, -1) : r
          );
          const newFoundations = [...foundationsArr];
          newFoundations[foundationIndex] = cardChar;
          nextStates.push(newRows.join("|") + "_" + newFoundations.join("_"));
        }
      });

      // if ace, start new foundation
      if (getSolverValue(cardChar) === 0) {
        const newRows = rows.map((r, i) =>
          i === sourceRowIndex ? r.slice(0, -1) : r
        );
        const newFoundations = [...foundationsArr, cardChar];
        nextStates.push(newRows.join("|") + "_" + newFoundations.join("_"));
      }
    });

    return nextStates;
  }

  const visited = new Set();
  const pathParent = new Map();
  const gameStack = [stateToSolver(streets)];
  visited.add(gameStack[0]);
  let iterations = 0;
  let lastTime = performance.now();

  while (gameStack.length > 0) {
    iterations++;
    if (iterations % 40000 === 0) {
      const now = performance.now();
      const elapsed = now - lastTime;
      console.log(
        `Iterations: ${iterations}, Stack: ${gameStack.length}, Visited: ${
          visited.size
        }, Time: ${elapsed.toFixed(2)}ms`
      );
      console.table(solverToState(gameStack[gameStack.length - 1]).rows);
      lastTime = now;
    }

    const solverStr = gameStack.pop();
    const [rowsStr] = solverStr.split("_");
    const emptyRows = rowsStr
      .split("|")
      .filter((row) => row.length === 0).length;

    if (emptyRows >= 2) {
      const path = [];
      let current = solverStr;
      while (current) {
        path.unshift(solverToState(current));
        current = pathParent.get(current);
      }
      console.log("probably winnable");
      return path;
    }

    const nextStates = nextSolverStates(solverStr);
    nextStates.forEach((nextStr) => {
      if (!visited.has(nextStr)) {
        visited.add(nextStr);
        pathParent.set(nextStr, solverStr);
        gameStack.push(nextStr);
      }
    });
  }
  console.log("not winnable");
  return false;
}

// Initialize the game
reset();
console.log(nextStates(stateToCompact(streets)));
isProbablyWinnable();
