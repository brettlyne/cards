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
  .card-wrapper {
    position: relative;
    width: 100vw;
    max-width: 12vw;
    max-height: 21vh;
    aspect-ratio: 2.5/3.5;
    filter: drop-shadow(0 0 0.4rem #4c350660);
    border-radius: 1em;
  }
</style>
