<script>
  import Card from "./lib/Card.svelte";
  import "./app.css";
  import "./snes.css";
  import { streets } from "./streetsAndAlleys.svelte.js";

  const themes = ["light", "dark", "neon", "dracula", "solarized", "sepia"];
  let currentTheme = "light";

  function setTheme(theme) {
    document.documentElement.className =
      theme === "light" ? "" : `theme-${theme}`;
    currentTheme = theme;
  }
</script>

<main>
  <header>
    <ul>
      <li><a class="snes-link text-ocean-color">Can I win?</a></li>
      <li><a class="snes-link text-ocean-color">Undo</a></li>
    </ul>
    <label class="snes-link text-ocean-color"
      >Theme&nbsp; <select
        bind:value={currentTheme}
        on:change={() => setTheme(currentTheme)}
        class="snes-link text-ocean-color"
      >
        {#each themes as theme}
          <option value={theme}>{theme}</option>
        {/each}
      </select>
    </label>
    <ul>
      <li><a class="snes-link text-ocean-color">New Game</a></li>
    </ul>
  </header>
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
  header {
    display: flex;
    justify-content: space-between;
    margin: auto;
    max-width: 1000px;
    padding-top: 22px;
    padding-bottom: 6px;

    ul {
      display: flex;
      list-style: none;
      padding: 0;
      margin: 0;
    }

    li {
      margin-right: 32px;
    }

    li:last-child {
      margin-right: 0;
    }

    select {
      padding: 6px 4px 4px;
    }
  }

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
