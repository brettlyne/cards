// Compact card representation for solver only (charCode 48-111)
export function cardToSolver(card) {
  const value = "A23456789TJQK".indexOf(card[0]); // 0-12
  const suit = card[1];
  // Leave 16 slots per suit (13 cards + 3 gap)
  if (suit === "H") return String.fromCharCode(48 + value);
  if (suit === "D") return String.fromCharCode(64 + value);
  if (suit === "C") return String.fromCharCode(80 + value);
  if (suit === "S") return String.fromCharCode(96 + value);
}

export function solverToCard(char) {
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

export function getSolverValue(char) {
  return (char.charCodeAt(0) - 48) % 16;
}

export function getSolverSuit(char) {
  const code = char.charCodeAt(0);
  if (code < 64) return 0; // Hearts
  if (code < 80) return 1; // Diamonds
  if (code < 96) return 2; // Clubs
  return 3; // Spades
}

export function canMoveSolver(card, targetCard) {
  if (!targetCard) return true;
  return getSolverValue(targetCard) === getSolverValue(card) + 1;
}

export function canMoveToFoundationSolver(card, topCard) {
  if (!topCard) {
    return getSolverValue(card) === 0; // Ace
  }
  return (
    getSolverSuit(card) === getSolverSuit(topCard) &&
    getSolverValue(card) === getSolverValue(topCard) + 1
  );
}

export function nextSolverStates(solverStr) {
  const [rowsStr, foundationsStr] = solverStr.split("_");
  const rows = rowsStr.split("|");
  const foundations = (foundationsStr || "").split("_");

  const nextStates = new Set();

  // Try moving to foundations
  rows.forEach((row, fromRow) => {
    if (row.length === 0) return;

    const card = row[row.length - 1];
    foundations.forEach((foundation, suit) => {
      if (canMoveToFoundationSolver(card, foundation?.[foundation.length - 1])) {
        const newRows = [...rows];
        const newFoundations = [...foundations];
        newRows[fromRow] = row.slice(0, -1);
        newFoundations[suit] = (foundation || "") + card;
        nextStates.add(
          newRows.join("|") + "_" + newFoundations.filter(Boolean).join("_")
        );
      }
    });
  });

  // Try moving between rows
  rows.forEach((fromRow, fromIndex) => {
    if (fromRow.length === 0) return;

    rows.forEach((toRow, toIndex) => {
      if (fromIndex === toIndex) return;

      const card = fromRow[fromRow.length - 1];
      const targetCard = toRow.length > 0 ? toRow[toRow.length - 1] : null;

      if (canMoveSolver(card, targetCard)) {
        const newRows = [...rows];
        newRows[fromIndex] = fromRow.slice(0, -1);
        newRows[toIndex] = toRow + card;
        nextStates.add(newRows.join("|") + "_" + foundations.filter(Boolean).join("_"));
      }
    });
  });

  return Array.from(nextStates);
}

export function getInitialMoves(initialState) {
  return nextSolverStates(initialState);
}
