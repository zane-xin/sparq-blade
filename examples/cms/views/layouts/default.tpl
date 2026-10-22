<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <title>{{ Page.Title }} - {{ Site.Name }}</title>
  <style>
    * { box-sizing: border-box; }
    body { margin: 0; color: #1f2933; background: #f7f9fc; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; line-height: 1.65; }
    a { color: #1d4ed8; text-decoration: none; }
    a:hover { text-decoration: underline; }
    .shell { display: grid; grid-template-columns: 280px minmax(0, 1fr); min-height: 100vh; }
    .sidebar { background: #ffffff; border-right: 1px solid #d9e2ec; padding: 28px 22px; position: sticky; top: 0; height: 100vh; overflow: auto; }
    .brand { display: block; color: #102a43; font-weight: 700; font-size: 20px; margin-bottom: 24px; }
    .nav-list, .submenu { list-style: none; margin: 0; padding: 0; }
    .nav-list > li { margin: 12px 0; }
    .nav-list a { display: inline-flex; align-items: center; min-height: 28px; color: #334e68; font-weight: 600; }
    .submenu { margin: 4px 0 0 14px; border-left: 2px solid #e4e7eb; padding-left: 12px; }
    .submenu li { margin: 4px 0; }
    .submenu a { color: #52606d; font-size: 14px; font-weight: 500; }
    .content { max-width: 960px; padding: 36px 44px 64px; }
    .crumb { color: #627d98; font-size: 14px; margin-bottom: 18px; }
    .crumb a { color: #627d98; margin-right: 8px; }
    .page-title { margin: 0; color: #102a43; font-size: 32px; line-height: 1.2; }
    .summary { max-width: 760px; color: #52606d; margin: 12px 0 30px; }
    section { background: #ffffff; border: 1px solid #d9e2ec; border-radius: 8px; padding: 22px 24px; margin: 18px 0; }
    h2 { margin: 0 0 12px; color: #102a43; font-size: 21px; }
    h3 { margin: 18px 0 8px; color: #243b53; }
    p { margin: 10px 0; }
    pre { margin: 14px 0; background: #102a43; color: #f0f4f8; border-radius: 6px; padding: 16px; overflow: auto; font-size: 14px; }
    code { background: #eef2f7; border-radius: 4px; padding: 1px 5px; }
    pre code, pre { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
    table { width: 100%; border-collapse: collapse; margin: 14px 0; background: #fff; }
    th, td { border-bottom: 1px solid #d9e2ec; padding: 10px 12px; text-align: left; vertical-align: top; }
    th { color: #243b53; background: #f0f4f8; font-weight: 700; }
    td:first-child { white-space: nowrap; }
    .note { background: #eef6ff; border-left: 4px solid #2563eb; color: #334e68; padding: 12px 16px; border-radius: 4px; }
    .demo-list { padding-left: 22px; }
    .demo-list li { margin: 10px 0; }
    .meta { color: #627d98; font-size: 14px; }
    .badge { display: inline-flex; align-items: center; min-height: 24px; padding: 0 9px; border-radius: 999px; background: #e0f2fe; color: #075985; font-size: 13px; font-weight: 700; }
    .doc-card { border: 1px solid #bcccdc; border-radius: 8px; padding: 16px; background: #f8fafc; }
    .doc-card h3 { margin-top: 0; }
    .doc-note { border-left: 4px solid #64748b; background: #f8fafc; padding: 12px 16px; border-radius: 4px; }
    .row-item { padding: 10px 0; border-bottom: 1px solid #e4e7eb; }
    .directive-panel { background: #fff; border: 1px solid #cbd5e1; border-radius: 8px; padding: 18px; margin: 14px 0; }
    .directive-panel::before { content: attr(data-kind); display: inline-block; margin-bottom: 10px; color: #475569; font-size: 13px; font-weight: 700; }
    .pager { display: flex; gap: 8px; align-items: center; padding: 12px 0; }
    .pager-theme { color: #102a43; font-size: 18px; }
    @@media (max-width: 820px) {
      .shell { grid-template-columns: 1fr; }
      .sidebar { position: static; height: auto; }
      .content { padding: 24px 18px 40px; }
    }
  </style>
</head>
<body>
  <div class="shell">
  <aside class="sidebar">
    <a class="brand" href="/">{{ Site.Name }}</a>
    <nav>
      <ul class="nav-list">
      @sq:nav(as=item)
        <li>
          <a href="{{ item.url }}">{{ item.title }}</a>
          @children(as=child, wrap="ul", class="submenu")
            <li><a href="{{ child.url }}">{{ child.title }}</a></li>
          @empty
          @end
        </li>
      @end
      </ul>
    </nav>
  </aside>

  <main class="content">
    @sq:crumb()
    <h1 class="page-title">{{ Page.Title }}</h1>
    <p class="summary">{{ Page.Summary }}</p>
    @yield('content')
  </main>
  </div>
</body>
</html>
