<article class="doc-card" data-kind="{{ kind }}">
  <h3>{{ title }}</h3>
  <div>{{ slot }}</div>
  @if(footer)
    <footer class="meta">{{ footer }}</footer>
  @end
</article>
