<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>@sq:title</title>
  @sq:seo
  <style>
    body { font-family: Arial, sans-serif; margin: 0; color: #1f2937; background: #f7f8fa; }
    header { background: #111827; color: white; padding: 20px 32px; }
    header h1 { margin: 0 0 6px; font-size: 24px; }
    header p { margin: 0; color: #cbd5e1; }
    .shell { display: grid; grid-template-columns: 260px 1fr; min-height: calc(100vh - 86px); }
    nav { background: white; border-right: 1px solid #e5e7eb; padding: 18px; }
    nav a { display: block; padding: 10px 12px; color: #374151; text-decoration: none; border-radius: 6px; }
    nav a.active { background: #e0f2fe; color: #075985; font-weight: 700; }
    main { padding: 28px 36px; max-width: 980px; }
    section { margin-bottom: 28px; }
    h2 { margin: 0 0 12px; font-size: 22px; }
    h3 { margin: 0 0 8px; font-size: 18px; }
    p { line-height: 1.7; }
    ul, ol { line-height: 1.8; }
    code, pre { font-family: Consolas, monospace; }
    pre { background: #0f172a; color: #e5e7eb; padding: 14px; border-radius: 6px; overflow: auto; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 14px; }
    .card { background: white; border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px; }
    .muted { color: #6b7280; }
    .badge { display: inline-block; padding: 2px 8px; background: #dcfce7; color: #166534; border-radius: 999px; font-size: 12px; }
  </style>
</head>
<body>
  <header>
    <h1>{{ Site.Name }}</h1>
    <p>{{ Meta.Description }}</p>
  </header>

  <div class="shell">
    @include('partials.nav')
    <main>
      @yield('content')
    </main>
  </div>
</body>
</html>
