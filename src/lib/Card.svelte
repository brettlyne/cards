<script>
  export let card = "2J";
</script>

<img class="card" src="/cards/{card}.svg" alt={card} />

<style>
  img {
    position: relative;
    max-width: 12vw;
    max-height: 21vh;
    aspect-ratio: 2.5/3.5;
    /* filter: invert(0.85); */
    filter: drop-shadow(0 0 0.4rem #4c350660);
  }
</style>
