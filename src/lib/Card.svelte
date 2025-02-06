<script>
  export let card = "2J";
  // dynamically import the correct card component
  $: cardComponent = async () => {
    const module = await import(`./cards/${card}.svelte`);
    return module.default;
  };
</script>

<div class="card-wrapper">
  {#await cardComponent() then Component}
    <svelte:component this={Component} />
  {/await}
</div>

<style>
  :root {
    --card-aspect-ratio: 2.5/3.5;
    --card-max-height: calc((92vh - 60px) / 4);
    --card-width: min(12vw, calc(var(--card-max-height) * (2.5 / 3.5)));
  }

  .card-wrapper {
    position: relative;
    width: var(--card-width);
    max-height: var(--card-max-height);
    aspect-ratio: var(--card-aspect-ratio);
    filter: drop-shadow(0 0 0.4rem #4c350660);
    border-radius: 1em;
  }
</style>
