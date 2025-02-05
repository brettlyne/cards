<script>
  import Card from "./lib/Card.svelte";
  import "./app.css";
  import "./snes.css";
  import { streets } from "./streetsAndAlleys.svelte.js";
</script>

<main>
  <div class="streets-and-alleys">
    {#each Array(4) as _, i}
      <div class="row left">
        {#each streets.rows[i * 2] as card}
          <Card {card} />
        {/each}
      </div>
      <Card />
      <div class="row right">
        {#each streets.rows[i * 2 + 1] as card}
          <Card {card} />
        {/each}
      </div>
    {/each}
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
