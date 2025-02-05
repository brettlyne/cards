<script>
  import Card from "./lib/Card.svelte";
  import { cardDeck, shuffle, dealStreetsAlleys } from "./utils/utils";
  import "./snes.css";

  let streetsAlleys = dealStreetsAlleys(shuffle(cardDeck));
  const rows = streetsAlleys.split("|");
  // separate each row into an array of 2-character strings
  // const cardRows = rows.map((row) => row.match(/.{2}/g));
  const cardRows = shuffle(cardDeck);
  console.log("🚀 ~ file: App.svelte:9 ~ cardRows:", cardRows);
</script>

<main>
  <div class="streets-and-alleys">
    <div class="row left">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
    </div>
    <div class="foundation">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
    </div>
    <div class="row right">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
    </div>

    <div class="row left">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
    </div>
    <div class="foundation">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
    </div>
    <div class="row right">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
    </div>
    <div class="row left">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
    </div>
    <div class="foundation">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
    </div>
    <div class="row right">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
    </div>
    <div class="row left">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
    </div>
    <div class="foundation">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
    </div>
    <div class="row right">
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
      <Card card={cardRows[Math.floor(Math.random() * 52)]} />
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
