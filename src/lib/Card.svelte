<script>
  export let card;
</script>

<img src="/cards/{card}.svg" alt={card} />

<style>
  img {
    position: relative;
    max-width: 14vw;
    max-height: 21vh;
    aspect-ratio: 2.5/3.5;
  }
</style>
