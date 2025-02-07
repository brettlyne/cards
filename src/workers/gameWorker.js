// Import utility functions
import { nextSolverStates } from "../utils/solver.js";
import { getQueueForWorker } from "../utils/queues.js";

// Initialize shared variables
const pathParent = new Map();
let gameQueue; // Will be initialized with specific strategy in init
let iterations = 0;
let lastProgressTime = Date.now();
let workerCount = 4; // Default, will be updated in init
let workerId; // Store worker ID at module level

// Define processQueue function at module scope
function processQueue(timeout = null) {
  const startTime = Date.now();

  while (gameQueue.length > 0) {
    // Get current timing info
    const now = Date.now();
    const interval = workerCount * 1000;
    const timeSinceStart = now - Math.floor(now / interval) * interval;
    const isMyTimeSlot = Math.floor(timeSinceStart / 1000) === workerId;

    // Check for timeout - stagger based on worker ID
    if (timeout && now - startTime > timeout + workerId * 1000) {
      self.postMessage({
        type: "log",
        data: {
          workerId,
          message: `Worker ${workerId} timeout after ${timeout + workerId * 1000}ms`
        }
      });
      sendPathParentData();
      return;
    }

    iterations++;
    const currentState = gameQueue.pop();

    // Report progress periodically when it's our time slot
    if (isMyTimeSlot && now - lastProgressTime > 900) {
      self.postMessage({
        type: "log",
        data: {
          workerId,
          message: `Worker ${workerId} progress: ${iterations} iterations, queue: ${gameQueue.length}`
        }
      });
      self.postMessage({
        type: "progress",
        data: {
          workerId,
          iterations,
          queueSize: gameQueue.length,
          pathParentSize: pathParent.size,
          currentState,
        },
      });
      lastProgressTime = now;
    }

    // Get next possible states
    const nextStates = nextSolverStates(currentState);

    // Check each next state
    for (const nextState of nextStates) {
      // Check if state is solved by examining foundations
      const [rowsStr, foundationsStr] = nextState.split("_");
      const rows = rowsStr.split("|");
      const foundations = (foundationsStr || "").split("_");

      const isComplete = rows.filter((row) => row.length === 0).length >= 3;

      if (isComplete) {
        // Found solution, reconstruct path and send back
        const path = [];
        let state = nextState;
        while (state !== null) {
          path.unshift(state);
          state = pathParent.get(state);
        }

        self.postMessage({
          type: "solution",
          data: { path },
        });
        return;
      }

      // Add unvisited states to queue
      if (!pathParent.has(nextState)) {
        gameQueue.push(nextState);
        pathParent.set(nextState, currentState);
      }
    }
  }

  // If we get here, we've run out of work
  sendPathParentData();
}

// Helper function to send path parent data
function sendPathParentData() {
  // Convert pathParent Map to array of entries for transfer
  const pathParentData = Array.from(pathParent.entries())
    .filter(([state]) => state && typeof state === 'string' && state.includes('_')) // Filter out invalid states
    .map(([state, parent]) => ({
      state,
      parent,
      score: gameQueue.getStateScore ? gameQueue.getStateScore(state) : null
    }));

  self.postMessage({
    type: "pathParentData",
    data: {
      workerId,
      iterations,
      pathParentSize: pathParent.size,
      pathParentData,
      queueSize: gameQueue.length,
      queueType: gameQueue.constructor.name,
    },
  });
}

// Worker message handler
self.onmessage = function (e) {
  const { type, data } = e.data;

  if (type === "init") {
    // Initial setup
    const {
      initialMoves,
      workerId: wid,
      workerCount: wc,
      allInitialMoves,
      timeout,
    } = data;
    workerCount = wc; // Update the worker count
    workerId = wid; // Update the worker ID
    self.postMessage({
      type: "log",
      data: {
        workerId,
        message: `Worker ${workerId} started with ${initialMoves.length} moves`
      }
    });

    // Initialize worker-specific variables
    iterations = 0;
    lastProgressTime = Date.now();
    pathParent.clear();
    gameQueue = getQueueForWorker(workerId);

    // First, populate pathParent with all initial states and their handlers
    allInitialMoves.forEach(({ move, handledBy }) => {
      pathParent.set(move, null);
    });

    // Add initial moves to queue
    initialMoves.forEach((move) => {
      gameQueue.push(move);
    });

    // Start processing if we have work
    if (gameQueue.length === 0) {
      sendPathParentData();
    } else {
      processQueue(timeout);
    }
  } else if (type === "getPathParentData") {
    // Convert pathParent Map to array of entries for transfer
    const pathParentData = Array.from(pathParent.entries())
      .filter(([state]) => state && typeof state === 'string' && state.includes('_')) // Filter out invalid states
      .map(([state, parent]) => ({
        state,
        parent,
        score: gameQueue.getStateScore ? gameQueue.getStateScore(state) : null
      }));

    self.postMessage({
      type: "pathParentData",
      data: {
        workerId,
        iterations,
        pathParentSize: pathParent.size,
        pathParentData,
        queueSize: gameQueue.length,
        queueType: gameQueue.constructor.name,
      },
    });
  }
};
