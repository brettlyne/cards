<script>
  import Card from "./lib/Card.svelte";
  import { cardDeck, shuffle, dealStreetsAlleys } from "./utils/utils";

  let streetsAlleys = dealStreetsAlleys(shuffle(cardDeck));
  const rows = streetsAlleys.split("|");
  // separate each row into an array of 2-character strings
  const cardRows = rows.map((row) => row.match(/.{2}/g));
</script>

<main>
  <h1>Cards Test</h1>

  <div class="card-area">
    <div>
      {#each cardRows.slice(0, 4) as row}
        <div class="card-row">
          {#each row as card}
            <Card {card} />
          {/each}
        </div>
      {/each}
    </div>

    <div>
      {#each cardRows.slice(4, 8) as row}
        <div class="card-row">
          {#each row as card}
            <Card {card} />
          {/each}
        </div>
      {/each}
    </div>
  </div>
</main>

<style>
  .card-area {
    display: flex;
    flex-direction: row;
    justify-content: space-around;
    width: 1000px;
    overflow: hidden;
  }
  .card-row {
    display: flex;
  }
  .card-row :global(*):not(:first-child) {
    margin-left: -80px;
  }
</style>
