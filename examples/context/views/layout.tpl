<!doctype html>
<html>
<head>
  <title>@sq:title</title>
  @sq:seo
</head>
<body>
  <header>
    <strong>{{ Site.Name }}</strong>
    <span>{{ Route.Path }}</span>
    <span>{{ Route.Alias }}</span>
  </header>

  <main>
    @yield('content')
  </main>
</body>
</html>
