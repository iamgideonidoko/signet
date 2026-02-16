package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/iamgideonidoko/signet/internal/middleware"
	"github.com/iamgideonidoko/signet/internal/models"
	"github.com/iamgideonidoko/signet/internal/services"
	"github.com/iamgideonidoko/signet/pkg/cache"
	"github.com/iamgideonidoko/signet/pkg/logger"
	"github.com/iamgideonidoko/signet/pkg/similarity"
	"github.com/iamgideonidoko/signet/pkg/validator"
)

type Handler struct {
	identService *services.IdentificationService
	cache        *cache.Cache
}

func NewHandler(identService *services.IdentificationService, cache *cache.Cache) *Handler {
	return &Handler{
		identService: identService,
		cache:        cache,
	}
}

// Identify handles POST /v1/identify.
func (h *Handler) Identify(c *fiber.Ctx) error {
	requestID := uuid.New().String()
	log := logger.WithField("request_id", requestID)

	var req models.IdentifyRequest

	if err := c.BodyParser(&req); err != nil {
		log.Warn("Failed to parse request body", map[string]any{
			"error": err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":      "Invalid request body",
			"request_id": requestID,
		})
	}

	// Validate request
	if err := validator.ValidateIdentifyRequest(req); err != nil {
		log.Warn("Request validation failed", map[string]any{
			"error": err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":      err.Error(),
			"request_id": requestID,
		})
	}

	// Set IP address from request
	req.IPAddress = middleware.AnonymizeIP(c.IP())

	// Compute hardware hash and set in context for rate limiting
	hardwareHash := similarity.ComputeHardwareHash(req.Signals)
	c.Locals("hardware_hash", hardwareHash)

	// Perform identification
	result, err := h.identService.Identify(c.Context(), req)
	if err != nil {
		log.Error("Identification failed", map[string]any{
			"error": err.Error(),
			"ip":    req.IPAddress,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":      "Failed to identify visitor",
			"request_id": requestID,
		})
	}

	// Track metrics
	_ = h.cache.IncrementMetric(c.Context(), "total_identifications")

	log.Info("Identification successful", map[string]any{
		"visitor_id": result.VisitorID,
		"is_new":     result.IsNew,
		"confidence": result.Confidence,
	})

	return c.Status(fiber.StatusOK).JSON(result)
}

// Health handles GET /health.
func (h *Handler) Health(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "healthy",
		"service": "signet-api",
	})
}

// Metrics handles GET /metrics.
func (h *Handler) Metrics(c *fiber.Ctx) error {
	ctx := c.Context()

	totalIdents, _ := h.cache.GetMetric(ctx, "total_identifications")
	newVisitors, _ := h.cache.GetMetric(ctx, "new_visitors")
	healedIdents, _ := h.cache.GetMetric(ctx, "healed_identifications")
	cacheHits, _ := h.cache.GetMetric(ctx, "cache_hits")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"total_identifications":  totalIdents,
		"new_visitors":           newVisitors,
		"healed_identifications": healedIdents,
		"cache_hits":             cacheHits,
		"cache_hit_rate":         calculateRate(cacheHits, totalIdents),
	})
}

// Analytics handles GET /api/analytics.
func (h *Handler) Analytics(c *fiber.Ctx) error {
	days := c.QueryInt("days", 7)
	if days > 90 {
		days = 90 // Cap at 90 days
	}

	analytics, err := h.identService.GetAnalytics(c.Context(), days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch analytics",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"analytics": analytics,
	})
}

// RecentIdentifications handles GET /api/identifications.
func (h *Handler) RecentIdentifications(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	if limit > 100 {
		limit = 100
	}

	identifications, err := h.identService.GetRecentIdentifications(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch identifications",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"identifications": identifications,
		"limit":           limit,
		"offset":          offset,
	})
}

// Dashboard serves the analytics dashboard HTML.
func (h *Handler) Dashboard(c *fiber.Ctx) error {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Signet Dashboard</title>
    <style>
        :root {
            --bg-primary: #000000;
            --bg-secondary: #0a0a0a;
            --bg-tertiary: #111111;
            --border-primary: #1a1a1a;
            --border-secondary: #2a2a2a;
            --text-primary: #ffffff;
            --text-secondary: #a1a1a1;
            --text-tertiary: #666666;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }

        @keyframes pulse {
            0%, 100% { opacity: 0.4; }
            50% { opacity: 1; }
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Inter', sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            padding: 2rem;
            min-height: 100vh;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            animation: fadeIn 0.5s ease-out;
        }

        .header {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            margin-bottom: 3rem;
            padding-bottom: 1.5rem;
            border-bottom: 1px solid var(--border-primary);
        }

        h1 {
            font-size: 2rem;
            font-weight: 700;
            color: var(--text-primary);
            letter-spacing: -0.04em;
        }

        .metrics {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 1rem;
            margin-bottom: 3rem;
        }

        .metric-card {
            background: var(--bg-secondary);
            border: 1px solid var(--border-primary);
            padding: 1.75rem;
            border-radius: 8px;
            position: relative;
            overflow: hidden;
            transition: all 0.2s ease;
        }

        .metric-card::before {
            content: "";
            position: absolute;
            inset: 0;
            border-radius: 8px;
            padding: 1px;
            background: linear-gradient(135deg, rgba(255,255,255,0.1), transparent);
            -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
            -webkit-mask-composite: xor;
            mask-composite: exclude;
            pointer-events: none;
        }

        .metric-card:hover {
            background: rgba(255,255,255,0.02);
            border-color: var(--border-secondary);
        }

        .metric-label {
            color: var(--text-tertiary);
            font-size: 0.75rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 0.75rem;
            font-weight: 500;
        }

        .metric-value {
            font-size: 2.5rem;
            font-weight: 700;
            color: var(--text-primary);
            letter-spacing: -0.02em;
            font-variant-numeric: tabular-nums;
        }

        .section-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 1.5rem;
        }

        h2 {
            font-size: 1.25rem;
            font-weight: 600;
            color: var(--text-primary);
            letter-spacing: -0.02em;
        }

        .auto-refresh {
            color: var(--text-tertiary);
            font-size: 0.875rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .pulse-dot {
            width: 6px;
            height: 6px;
            background: var(--text-tertiary);
            border-radius: 50%;
            animation: pulse 2s ease-in-out infinite;
        }

        .table-wrapper {
            background: var(--bg-secondary);
            border: 1px solid var(--border-primary);
            border-radius: 8px;
            overflow: hidden;
        }

        table {
            width: 100%;
            border-collapse: collapse;
        }

        th, td {
            padding: 1rem 1.25rem;
            text-align: left;
        }

        th {
            background: var(--bg-tertiary);
            color: var(--text-tertiary);
            font-weight: 500;
            font-size: 0.75rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            border-bottom: 1px solid var(--border-primary);
        }

        td {
            border-bottom: 1px solid var(--border-primary);
            font-size: 0.9375rem;
        }

        tbody tr {
            transition: background 0.15s ease;
        }

        tbody tr:hover {
            background: rgba(255,255,255,0.02);
        }

        tbody tr:last-child td {
            border-bottom: none;
        }

        .visitor-id {
            font-family: 'SF Mono', 'Monaco', 'Inconsolata', monospace;
            font-size: 0.875rem;
            color: var(--text-secondary);
        }

        .badge {
            display: inline-flex;
            align-items: center;
            padding: 0.25rem 0.625rem;
            border-radius: 4px;
            font-size: 0.75rem;
            font-weight: 600;
            letter-spacing: 0.02em;
            border: 1px solid;
        }

        .badge-new {
            background: rgba(255,255,255,0.1);
            color: var(--text-primary);
            border-color: rgba(255,255,255,0.2);
        }

        .badge-returning {
            background: rgba(255,255,255,0.05);
            color: var(--text-secondary);
            border-color: var(--border-secondary);
        }

        .confidence {
            font-weight: 600;
            font-variant-numeric: tabular-nums;
        }

        .confidence-high { color: var(--text-primary); }
        .confidence-medium { color: #e5e5e5; }
        .confidence-low { color: var(--text-secondary); }

        .timestamp {
            color: var(--text-secondary);
            font-size: 0.875rem;
        }

        @media (max-width: 768px) {
            body { padding: 1rem; }
            .metrics { grid-template-columns: 1fr; }
            .header { margin-bottom: 2rem; }
            h1 { font-size: 1.5rem; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Signet Dashboard</h1>
        </div>

        <div class="metrics" id="metrics">
            <div class="metric-card">
                <div class="metric-label">Total Identifications</div>
                <div class="metric-value" id="total">-</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">New Visitors</div>
                <div class="metric-value" id="new">-</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Healed Identifications</div>
                <div class="metric-value" id="healed">-</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Cache Hit Rate</div>
                <div class="metric-value" id="cache">-</div>
            </div>
        </div>

        <div class="section-header">
            <h2>Recent Identifications</h2>
            <div class="auto-refresh">
                <span class="pulse-dot"></span>
                <span>Auto-refresh every 5s</span>
            </div>
        </div>

        <div class="table-wrapper">
            <table>
                <thead>
                    <tr>
                        <th>Visitor ID</th>
                        <th>IP Address</th>
                        <th>Confidence</th>
                        <th>Status</th>
                        <th>Timestamp</th>
                    </tr>
                </thead>
                <tbody id="identifications"></tbody>
            </table>
        </div>
    </div>

    <script>
        async function loadMetrics() {
            try {
                const res = await fetch('/metrics');
                const data = await res.json();
                document.getElementById('total').textContent = (data.total_identifications || 0).toLocaleString();
                document.getElementById('new').textContent = (data.new_visitors || 0).toLocaleString();
                document.getElementById('healed').textContent = (data.healed_identifications || 0).toLocaleString();
                document.getElementById('cache').textContent = (data.cache_hit_rate || 0).toFixed(1) + '%';
            } catch (err) {
                console.error('Failed to load metrics:', err);
            }
        }

        async function loadIdentifications() {
            try {
                const res = await fetch('/api/identifications?limit=20');
                const data = await res.json();
                const tbody = document.getElementById('identifications');

                if (!data.identifications || data.identifications.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="5" style="text-align: center; color: var(--text-tertiary); padding: 2rem;">No identifications yet</td></tr>';
                    return;
                }

                tbody.innerHTML = data.identifications.map(i => {
                    const confidence = (i.confidence_score * 100).toFixed(1);
                    const confClass = confidence >= 90 ? 'high' : confidence >= 70 ? 'medium' : 'low';
                    const isNew = i.confidence_score === 1.0 && !i.is_bot;
                    const timestamp = new Date(i.created_at).toLocaleString();

                    return ` + "`" + `
                        <tr>
                            <td><span class="visitor-id">${i.visitor_id.substring(0, 12)}...</span></td>
                            <td>${i.ip_address}</td>
                            <td><span class="confidence confidence-${confClass}">${confidence}%</span></td>
                            <td><span class="badge badge-${isNew ? 'new' : 'returning'}">${isNew ? 'NEW' : 'RETURNING'}</span></td>
                            <td><span class="timestamp">${timestamp}</span></td>
                        </tr>
                    ` + "`" + `;
                }).join('');
            } catch (err) {
                console.error('Failed to load identifications:', err);
            }
        }

        loadMetrics();
        loadIdentifications();
        setInterval(() => {
            loadMetrics();
            loadIdentifications();
        }, 5000);
    </script>
</body>
</html>
	`

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

func calculateRate(numerator, denominator int64) float64 {
	if denominator == 0 {
		return 0.0
	}
	return (float64(numerator) / float64(denominator)) * 100
}
