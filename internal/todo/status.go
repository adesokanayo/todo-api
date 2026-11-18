package todo

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type StatusComponent struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type StatusReport struct {
	Overall      string            `json:"overall"`
	Message      string            `json:"message"`
	CheckedAt    string            `json:"checked_at"`
	Uptime       string            `json:"uptime"`
	UptimeSecond int64             `json:"uptime_seconds"`
	Components   []StatusComponent `json:"components"`
}

func (h *handler) collectStatus(ctx context.Context) StatusReport {
	dbErr := h.store.Ping(ctx)
	dbOK := dbErr == nil

	message := "All systems operational"
	overall := "operational"
	if !dbOK {
		message = "Service disruption — database unreachable"
		overall = "outage"
	}

	uptime := time.Since(h.startedAt).Round(time.Second)

	return StatusReport{
		Overall:      overall,
		Message:      message,
		CheckedAt:    time.Now().UTC().Format(time.RFC3339),
		Uptime:       formatDuration(uptime),
		UptimeSecond: int64(uptime.Seconds()),
		Components: []StatusComponent{
			{Name: "API", Status: "up", Detail: "Serving HTTP requests"},
			{Name: "PostgreSQL", Status: componentStatus(dbOK), Detail: dbDetail(dbOK, dbErr)},
		},
	}
}

func componentStatus(ok bool) string {
	if ok {
		return "up"
	}
	return "down"
}

func (h *handler) statusHealth(w http.ResponseWriter, r *http.Request) {
	report := h.collectStatus(r.Context())
	status := http.StatusOK
	if report.Overall != "operational" {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, report)
}

func (h *handler) statusPage(w http.ResponseWriter, r *http.Request) {
	report := h.collectStatus(r.Context())

	bannerClass := report.Overall
	overallLabel := "Operational"
	overallIcon := "✓"
	if report.Overall != "operational" {
		overallLabel = "Outage"
		overallIcon = "!"
	}

	rows := ""
	for _, c := range report.Components {
		rows += componentCard(c)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if report.Overall != "operational" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	fmt.Fprintf(w, statusHTML,
		bannerClass, overallLabel,
		bannerClass, overallIcon, report.Message,
		report.Uptime, report.CheckedAt,
		rows,
	)
}

func componentCard(c StatusComponent) string {
	up := c.Status == "up"
	statusLabel := "Operational"
	cardClass := "up"
	dotClass := "dot-up"
	if !up {
		statusLabel = "Down"
		cardClass = "down"
		dotClass = "dot-down"
	}
	return fmt.Sprintf(`<article class="service %s">
  <div class="service-head">
    <span class="dot %s" aria-hidden="true"></span>
    <h3>%s</h3>
    <span class="pill %s">%s</span>
  </div>
  <p class="service-detail">%s</p>
</article>`, cardClass, dotClass, c.Name, cardClass, statusLabel, c.Detail)
}

func dbDetail(ok bool, err error) string {
	if ok {
		return "Connected and responding to ping"
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
      --violet: #8b5cf6;
      --violet-soft: #a78bfa;
      --cyan: #22d3ee;
      --pink: #f472b6;
      --emerald: #34d399;
      --emerald-bg: rgba(52, 211, 153, 0.15);
      --rose: #fb7185;
      --rose-bg: rgba(251, 113, 133, 0.18);
      --surface: rgba(255, 255, 255, 0.08);
      --surface-strong: rgba(255, 255, 255, 0.12);
      --border: rgba(255, 255, 255, 0.14);
      --text: #f8fafc;
      --muted: #94a3b8;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      font-family: "Inter", system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
      color: var(--text);
      background: #0c0a14;
      background-image:
        radial-gradient(ellipse 80%% 50%% at 20%% -10%%, rgba(139, 92, 246, 0.45), transparent),
        radial-gradient(ellipse 60%% 40%% at 90%% 10%%, rgba(34, 211, 238, 0.25), transparent),
        radial-gradient(ellipse 50%% 30%% at 50%% 100%%, rgba(244, 114, 182, 0.2), transparent);
      line-height: 1.55;
    }
    .wrap {
      max-width: 720px;
      margin: 0 auto;
      padding: 2.5rem 1.25rem 4rem;
    }
    .brand {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      margin-bottom: 0.5rem;
    }
    .logo {
      width: 44px;
      height: 44px;
      border-radius: 14px;
      background: linear-gradient(135deg, var(--violet), var(--cyan));
      display: grid;
      place-items: center;
      font-weight: 800;
      font-size: 1.1rem;
      box-shadow: 0 8px 24px rgba(139, 92, 246, 0.35);
    }
    h1 {
      margin: 0;
      font-size: 1.75rem;
      font-weight: 700;
      letter-spacing: -0.03em;
      background: linear-gradient(90deg, #fff, var(--violet-soft));
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }
    .subtitle {
      color: var(--muted);
      margin: 0.35rem 0 2rem;
      font-size: 1rem;
    }
    .banner {
      display: flex;
      align-items: flex-start;
      gap: 1rem;
      padding: 1.25rem 1.35rem;
      border-radius: 16px;
      margin-bottom: 1.75rem;
      border: 1px solid var(--border);
      backdrop-filter: blur(12px);
    }
    .banner.operational {
      background: linear-gradient(135deg, var(--emerald-bg), rgba(34, 211, 238, 0.08));
      border-color: rgba(52, 211, 153, 0.35);
    }
    .banner.outage {
      background: linear-gradient(135deg, var(--rose-bg), rgba(139, 92, 246, 0.08));
      border-color: rgba(251, 113, 133, 0.4);
    }
    .banner-icon {
      width: 2.5rem;
      height: 2.5rem;
      border-radius: 12px;
      display: grid;
      place-items: center;
      font-weight: 700;
      font-size: 1.25rem;
      flex-shrink: 0;
    }
    .banner.operational .banner-icon {
      background: var(--emerald);
      color: #042f1a;
    }
    .banner.outage .banner-icon {
      background: var(--rose);
      color: #450a0a;
    }
    .banner h2 {
      margin: 0;
      font-size: 1.125rem;
      font-weight: 600;
    }
    .banner p {
      margin: 0.25rem 0 0;
      color: var(--muted);
      font-size: 0.9rem;
    }
    .stats {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
      gap: 0.85rem;
      margin-bottom: 1.75rem;
    }
    .stat {
      padding: 1rem 1.1rem;
      border-radius: 14px;
      background: var(--surface);
      border: 1px solid var(--border);
      backdrop-filter: blur(8px);
    }
    .stat-label {
      font-size: 0.7rem;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: var(--cyan);
      font-weight: 600;
      margin-bottom: 0.35rem;
    }
    .stat-value {
      font-size: 1.05rem;
      font-weight: 600;
      word-break: break-word;
    }
    .section-title {
      font-size: 0.75rem;
      text-transform: uppercase;
      letter-spacing: 0.1em;
      color: var(--pink);
      font-weight: 700;
      margin: 0 0 0.75rem 0.15rem;
    }
    .services {
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
      margin-bottom: 1.75rem;
    }
    .service {
      padding: 1.1rem 1.25rem;
      border-radius: 14px;
      background: var(--surface-strong);
      border: 1px solid var(--border);
      border-left-width: 4px;
    }
    .service.up { border-left-color: var(--emerald); }
    .service.down { border-left-color: var(--rose); }
    .service-head {
      display: flex;
      align-items: center;
      gap: 0.6rem;
      flex-wrap: wrap;
    }
    .service h3 {
      margin: 0;
      flex: 1;
      font-size: 1rem;
      font-weight: 600;
    }
    .dot {
      width: 8px;
      height: 8px;
      border-radius: 50%%;
    }
    .dot-up { background: var(--emerald); box-shadow: 0 0 10px var(--emerald); }
    .dot-down { background: var(--rose); box-shadow: 0 0 10px var(--rose); }
    .pill {
      font-size: 0.7rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.04em;
      padding: 0.25rem 0.65rem;
      border-radius: 999px;
    }
    .pill.up { background: var(--emerald-bg); color: var(--emerald); }
    .pill.down { background: var(--rose-bg); color: var(--rose); }
    .service-detail {
      margin: 0.5rem 0 0 1.4rem;
      color: var(--muted);
      font-size: 0.875rem;
    }
    .api-card {
      padding: 1.15rem 1.25rem;
      border-radius: 14px;
      background: linear-gradient(160deg, rgba(139, 92, 246, 0.2), rgba(12, 10, 20, 0.6));
      border: 1px solid rgba(167, 139, 250, 0.35);
      margin-bottom: 1.5rem;
    }
    .api-card h3 {
      margin: 0 0 0.5rem;
      font-size: 0.95rem;
      color: var(--violet-soft);
    }
    .api-card p {
      margin: 0 0 0.75rem;
      font-size: 0.875rem;
      color: var(--muted);
    }
    code.endpoint {
      display: block;
      padding: 0.65rem 0.85rem;
      border-radius: 10px;
      background: rgba(0, 0, 0, 0.35);
      font-family: ui-monospace, "SF Mono", Menlo, monospace;
      font-size: 0.8rem;
      color: var(--cyan);
      word-break: break-all;
    }
    .api-card a {
      color: var(--cyan);
      font-weight: 500;
    }
    footer {
      font-size: 0.8125rem;
      color: var(--muted);
      text-align: center;
    }
    footer a { color: var(--violet-soft); text-decoration: none; }
    footer a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <div class="wrap">
    <header class="brand">
      <div class="logo" aria-hidden="true">T</div>
      <div>
        <h1>Todo API</h1>
        <p class="subtitle">Live service status · <span class="pill %s" style="display:inline-block;vertical-align:middle;">%s</span></p>
      </div>
    </header>

    <div class="banner %s">
      <div class="banner-icon">%s</div>
      <div>
        <h2>%s</h2>
        <p>Components are monitored on each page load.</p>
      </div>
    </div>

    <div class="stats">
      <div class="stat">
        <div class="stat-label">Uptime</div>
        <div class="stat-value">%s</div>
      </div>
      <div class="stat">
        <div class="stat-label">Last check (UTC)</div>
        <div class="stat-value">%s</div>
      </div>
    </div>

    <p class="section-title">Components</p>
    <div class="services">%s</div>

    <div class="api-card">
      <h3>Status health API</h3>
      <p>Machine-readable report for monitors and integrations (same data as this page).</p>
      <code class="endpoint">GET /status/health</code>
      <p style="margin-top:0.75rem;margin-bottom:0"><a href="/status/health">Open JSON response →</a></p>
    </div>

    <footer>
      Refreshes every 60s · Liveness probe <a href="/health">/health</a>
    </footer>
  </div>
</body>
</html>
`