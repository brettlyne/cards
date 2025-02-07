// Base queue class with common functionality
class BaseQueue {
  constructor() {
    this.queues = Array(100)
      .fill(null)
      .map(() => []);
    this.size = 0;
  }

  push(state) {
    const score = Math.max(0, Math.min(99, Math.floor(this.getStateScore(state))));
    this.queues[score].push(state);
    this.size++;
  }

  pop() {
    for (let i = this.queues.length - 1; i >= 0; i--) {
      if (this.queues[i].length > 0) {
        this.size--;
        return this.queues[i].shift();
      }
    }
    return undefined;
  }

  get length() {
    return this.size;
  }

  // Helper methods for scoring
  getFoundationScore(foundations) {
    let totalFoundationCards = 0;
    foundations.forEach((f) => {
      if (f) totalFoundationCards += f.length;
    });
    return totalFoundationCards * 2; // 0-80 range
  }

  getAccessibilityScore(rows, foundations) {
    let accessibilityScore = 0;
    foundations.forEach((topCard, suit) => {
      if (!topCard || this.getSolverValue(topCard) < 12) {
        // If not K
        const nextValue = topCard ? this.getSolverValue(topCard) + 1 : 0;
        const nextCardChar = String.fromCharCode(nextValue * 4 + suit + 48);
        rows.forEach((row) => {
          const cardIndex = row.indexOf(nextCardChar);
          if (cardIndex !== -1) {
            const cardsOnTop = row.length - cardIndex - 1;
            accessibilityScore -= cardsOnTop * 2;
          }
        });
      }
    });
    return accessibilityScore; // -40 to 0 range
  }

  getSolverValue(char) {
    return (char.charCodeAt(0) - 48) % 16;
  }
}

// Strategy 1: Current strategy (foundation score only)
export class CurrentStrategyQueue extends BaseQueue {
  getStateScore(state) {
    const [rowsStr, foundationsStr] = state.split("_");
    const foundations = (foundationsStr || "").split("_");
    return this.getFoundationScore(foundations);
  }
}

// Strategy 2: Pure BFS (always score of 0 for FIFO behavior)
export class BFSQueue extends BaseQueue {
  getStateScore(state) {
    return 0; // Always return 0 for FIFO behavior
  }
}

// Strategy 3: Only accessibility score
export class AccessibilityQueue extends BaseQueue {
  getStateScore(state) {
    const [rowsStr, foundationsStr] = state.split("_");
    const rows = rowsStr.split("|");
    const foundations = (foundationsStr || "").split("_");
    const score = this.getAccessibilityScore(rows, foundations);
    // Normalize to 0-99 range by adding 40 (since score is -40 to 0) and scaling
    return score + 40;
  }
}

// Strategy 4: Combined score with 1:2 weighting (accessibility:foundations)
export class CombinedQueue extends BaseQueue {
  getStateScore(state) {
    const [rowsStr, foundationsStr] = state.split("_");
    const rows = rowsStr.split("|");
    const foundations = (foundationsStr || "").split("_");

    const foundationScore = this.getFoundationScore(foundations);
    const accessibilityScore = this.getAccessibilityScore(rows, foundations);

    // Normalize to 0-99 range:
    // accessibilityScore is -40 to 0, so add 40 to make it 0-40
    // foundationScore is 0-80
    // Total range should be 0-99
    const score = (accessibilityScore + 40) + Math.floor(foundationScore * 0.7);
    return Math.max(0, Math.min(99, score));
  }
}

// Factory function to get the appropriate queue based on worker ID
export function getQueueForWorker(workerId) {
  switch (workerId) {
    case 1:
      return new CurrentStrategyQueue();
    case 2:
      return new BFSQueue();
    case 3:
      return new AccessibilityQueue();
    case 4:
      return new CombinedQueue();
    default:
      return new CurrentStrategyQueue();
  }
}
