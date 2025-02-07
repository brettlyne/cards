import { cardDeck, shuffle } from "./utils/utils";
import { cardToSolver, solverToCard, getInitialMoves } from "./utils/solver";

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

function stateToCompact(state) {
  const { rows, foundations } = state;
  const rowsStr = rows
    .map((row) => row.map((card) => cardToSolver(card)).join(""))
    .join("|");
  const foundationsStr = foundations
    .map((foundation) => foundation.map((card) => cardToSolver(card)).join(""))
    .filter(Boolean)
    .join("_");
  return rowsStr + (foundationsStr ? "_" + foundationsStr : "");
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
  const deck = shuffle(cardDeck);
  streets.rows = dealStreetsAlleys(deck);
  streets.foundations = [];
}

export function isProbablyWinnable() {
  return new Promise((resolve) => {
    const NUM_WORKERS = 4;
    const workers = [];
    let foundSolution = false;
    let workersWithNoWork = new Set();

    // Get initial state and moves
    const initialState = stateToCompact(streets);
    const allMoves = getInitialMoves(initialState);
    console.log(`Total initial moves: ${allMoves.length}`);

    // Distribute moves among workers - ensure each worker gets at least one move if possible
    const activeWorkerCount = Math.min(NUM_WORKERS, allMoves.length);
    const baseMovesPerWorker = Math.floor(allMoves.length / activeWorkerCount);
    const extraMoves = allMoves.length % activeWorkerCount;

    // Create and start workers
    const workerPromises = [];
    for (let i = 0; i < NUM_WORKERS; i++) {
      const worker = new Worker(
        new URL("./workers/gameWorker.js", import.meta.url),
        { type: "module" }
      );
      workers.push(worker);

      // Calculate move range for this worker
      let start = 0;
      let end = 0;
      if (i < activeWorkerCount) {
        start = i * baseMovesPerWorker + Math.min(i, extraMoves);
        end = start + baseMovesPerWorker + (i < extraMoves ? 1 : 0);
      }

      // Create promise for this worker's completion
      const workerPromise = new Promise((resolve) => {
        worker.onmessage = (e) => {
          if (e.data.type === "pathParentData") {
            resolve(e.data.data);
          } else if (e.data.type === "progress") {
            // Handle progress updates
            const { iterations, queueSize, pathParentSize, currentState } =
              e.data.data;
            console.log(
              `Worker ${i} progress: ${iterations} iterations, queue: ${queueSize}, states: ${pathParentSize}`
            );
          } else if (e.data.type === "log") {
            // Handle log messages
            console.log(
              `[Worker ${e.data.data.workerId}] ${e.data.data.message}`
            );
          }
        };
      });
      workerPromises.push(workerPromise);

      // Send initial work to the worker
      worker.postMessage({
        type: "init",
        data: {
          workerId: i,
          workerCount: NUM_WORKERS,
          initialMoves: allMoves.slice(start, end),
          allInitialMoves: allMoves,
          timeout: 8000 + i * 2000, // Base timeout of 8 seconds (each worker will add its own stagger)
        },
      });
    }

    // Wait for all workers to complete or timeout
    Promise.all(workerPromises).then((workerData) => {
      // Analyze results
      const statesByWorker = workerData.map(
        (data) => new Set(data.pathParentData.map((entry) => entry.state))
      );

      // Analyze duplicate work
      const totalUniqueStates = new Set(
        workerData.flatMap((data) =>
          data.pathParentData.map((entry) => entry.state)
        )
      ).size;

      console.log("\nWorker Statistics at Timeout:");
      workerData.forEach((data) => {
        console.log(`\nWorker ${data.workerId} (${data.queueType}):`);
        console.log(`- Iterations: ${data.iterations}`);
        console.log(`- States explored: ${data.pathParentSize}`);
        console.log(`- Queue size remaining: ${data.queueSize}`);
      });

      console.log("\nDuplicate Work Analysis:");
      console.log(
        `Total unique states across all workers: ${totalUniqueStates}`
      );
      console.log(
        "States explored by each worker:",
        workerData.map((d) => d.pathParentSize)
      );

      // Calculate pairwise overlap
      for (let i = 0; i < workerData.length; i++) {
        for (let j = i + 1; j < workerData.length; j++) {
          const overlap = new Set(
            [...statesByWorker[i]].filter((x) => statesByWorker[j].has(x))
          ).size;
          console.log(
            `Overlap between Worker ${workerData[i].workerId} and ${workerData[j].workerId}: ${overlap} states ` +
              `(${Math.round(
                (overlap /
                  Math.min(statesByWorker[i].size, statesByWorker[j].size)) *
                  100
              )}%)`
          );
        }
      }

      workers.forEach((worker) => worker.terminate());
      resolve(false);
    });
  });
}

// Initialize the game
reset();
isProbablyWinnable();
