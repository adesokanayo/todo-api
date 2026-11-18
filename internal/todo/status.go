package todo

import (
	"fmt"
	"net/http"
	"time"
)

func (h *handler) statusPage(w http.ResponseWriter, r *http.Request) {
	dbErr := h.store.Ping(r.Context())
	dbOK := dbErr == nil
	overallOK := dbOK

	overallLabel := "All systems operational"
	overallClass := "operational"
	if !dbOK {
		overallLabel = "Service disruption"
		overallClass = "outage"
	}

	uptime := time.Since(h.startedAt).Round(time.Second)
	checked := time.Now().UTC().Format(time.RFC3339)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !overallOK {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	fmt.Fprintf(w, statusHTML,
		overallClass, overallLabel,
		componentRow("API", true, "Serving requests"),
		componentRow("PostgreSQL", dbOK, dbDetail(dbOK, dbErr)),
		formatDuration(uptime),
		checked,
	)
}

func componentRow(name string, ok bool, detail string) string {
	status := "Operational"
	chipClass := "up"
	if !ok {
		status = "Down"
		chipClass = "down"
	}
	return fmt.Sprintf(`<tr>
  <td class="component">%s</td>
  <td><span class="chip %s">%s</span></td>
  <td class="detail">%s</td>
</tr>`, name, chipClass, status, detail)
}

func dbDetail(ok bool, err error) string {
	if ok {
		return "Connected"
	}
	if err != nil {
		return "Connection failed"
	}
	return "Unavailable"
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h >= 24 {
		days := h / 24
		h = h % 24
		return fmt.Sprintf("%dd %dh %dm", days, h, m)
	}
	return fmt.Sprintf("%dh %dm %ds", h, m, s)
}

const statusHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="refresh" content="60">
  <title>Todo API — Status</title>
  <style>
    :root {
      color-scheme: light dark;
      --bg: #0f1419;
      --surface: #1a2332;
      --border: #2d3a4d;
      --text: #e7ecf3;
      --muted: #8b9cb3;
      --green: #3dd68c;
      --green-dim: #1e3d32;
      --red: #f87171;
      --red-dim: #3d2222;
      --amber: #fbbf24;
      --amber-dim: #3d3422;
    }
    @media (prefers-color-scheme: light) {
      :root {
        --bg: #f4f6f9;
        --surface: #ffffff;
        --border: #e2e8f0;
        --text: #1e293b;
        --muted: #64748b;
        --green-dim: #dcfce7;
        --red-dim: #fee2e2;
        --amber-dim: #fef3c7;
      }
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      font-family: system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
      background: var(--bg);
      color: var(--text);
      line-height: 1.5;
    }
    main {
      max-width: 640px;
      margin: 0 auto;
      padding: 2.5rem 1.25rem 3rem;
    }
    h1 {
      font-size: 1.5rem;
      font-weight: 600;
      margin: 0 0 0.25rem;
      letter-spacing: -0.02em;
    }
    .subtitle {
      color: var(--muted);
      font-size: 0.9375rem;
      margin-bottom: 2rem;
    }
    .banner {
      border-radius: 12px;
      padding: 1rem 1.25rem;
      margin-bottom: 1.5rem;
      font-weight: 600;
      display: flex;
      align-items: center;
      gap: 0.625rem;
    }
    .banner::before {
      content: '';
      width: 10px;
      height: 10px;
      border-radius: 50%%;
      flex-shrink: 0;
    }
    .banner.operational {
      background: var(--green-dim);
      color: var(--green);
    }
    .banner.operational::before { background: var(--green); }
    .banner.outage {
      background: var(--red-dim);
      color: var(--red);
    }
    .banner.outage::before { background: var(--red); }
    .card {
      background: var(--surface);
      border: 1px solid var(--border);
      border-radius: 12px;
      overflow: hidden;
      margin-bottom: 1.5rem;
    }
    table {
      width: 100%%;
      border-collapse: collapse;
      font-size: 0.9375rem;
    }
    th, td {
      padding: 0.875rem 1.25rem;
      text-align: left;
      border-bottom: 1px solid var(--border);
    }
    tr:last-child td { border-bottom: none; }
    th {
      font-size: 0.75rem;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      color: var(--muted);
      font-weight: 600;
    }
    .component { font-weight: 500; }
    .detail { color: var(--muted); font-size: 0.875rem; }
    .chip {
      display: inline-block;
      font-size: 0.75rem;
      font-weight: 600;
      padding: 0.2rem 0.55rem;
      border-radius: 999px;
    }
    .chip.up { background: var(--green-dim); color: var(--green); }
    .chip.down { background: var(--red-dim); color: var(--red); }
    .meta {
      display: flex;
      flex-wrap: wrap;
      gap: 1.25rem;
      font-size: 0.875rem;
      color: var(--muted);
    }
    .meta strong { color: var(--text); font-weight: 500; }
    footer {
      margin-top: 2rem;
      font-size: 0.8125rem;
      color: var(--muted);
    }
    footer a { color: var(--muted); }
    footer a:hover { color: var(--text); }
  </style>
</head>
<body>
  <main>
    <h1>Todo API</h1>
    <p class="subtitle">System status and component health</p>
    <div class="banner %s">%s</div>
    <div class="card">
      <table>
        <thead>
          <tr>
            <th>Component</th>
            <th>Status</th>
            <th>Detail</th>
          </tr>
        </thead>
        <tbody>
          %s
          %s
        </tbody>
      </table>
    </div>
    <div class="meta">
      <span>Process uptime: <strong>%s</strong></span>
      <span>Last checked: <strong>%s</strong></span>
    </div>
    <footer>
      Auto-refreshes every 60s · JSON health at <a href="/health">/health</a>
    </footer>
  </main>
</body>
</html>
`