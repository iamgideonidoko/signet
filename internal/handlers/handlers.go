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
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            background: #0f0f23;
            color: #e0e0e0;
            padding: 2rem;
        }
        .container { max-width: 1400px; margin: 0 auto; }
        h1 { margin-bottom: 2rem; color: #00d9ff; }
        .metrics { 
            display: grid; 
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }
        .metric-card {
            background: #1a1a2e;
            border: 1px solid #2a2a3e;
            padding: 1.5rem;
            border-radius: 8px;
        }
        .metric-value { font-size: 2rem; font-weight: bold; color: #00d9ff; }
        .metric-label { color: #888; margin-top: 0.5rem; }
        table { 
            width: 100%; 
            border-collapse: collapse;
            background: #1a1a2e;
            border-radius: 8px;
            overflow: hidden;
        }
        th, td { padding: 1rem; text-align: left; border-bottom: 1px solid #2a2a3e; }
        th { background: #252540; color: #00d9ff; font-weight: 600; }
        tr:hover { background: #252540; }
        .badge { 
            display: inline-block;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-size: 0.875rem;
        }
        .badge-new { background: #00d9ff; color: #0f0f23; }
        .badge-returning { background: #4a4a6a; color: #e0e0e0; }
        .confidence { font-weight: bold; }
        .confidence-high { color: #00ff88; }
        .confidence-medium { color: #ffaa00; }
        .confidence-low { color: #ff4444; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Signet Dashboard</h1>
        
        <div class="metrics" id="metrics">
            <div class="metric-card">
                <div class="metric-value" id="total">-</div>
                <div class="metric-label">Total Identifications</div>
            </div>
            <div class="metric-card">
                <div class="metric-value" id="new">-</div>
                <div class="metric-label">New Visitors</div>
            </div>
            <div class="metric-card">
                <div class="metric-value" id="healed">-</div>
                <div class="metric-label">Healed Identifications</div>
            </div>
            <div class="metric-card">
                <div class="metric-value" id="cache">-</div>
                <div class="metric-label">Cache Hit Rate</div>
            </div>
        </div>

        <h2 style="margin-bottom: 1rem;">Recent Identifications</h2>
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

    <script>
        async function loadMetrics() {
            const res = await fetch('/metrics');
            const data = await res.json();
            document.getElementById('total').textContent = data.total_identifications || 0;
            document.getElementById('new').textContent = data.new_visitors || 0;
            document.getElementById('healed').textContent = data.healed_identifications || 0;
            document.getElementById('cache').textContent = (data.cache_hit_rate || 0).toFixed(1) + '%';
        }

        async function loadIdentifications() {
            const res = await fetch('/api/identifications?limit=20');
            const data = await res.json();
            const tbody = document.getElementById('identifications');
            tbody.innerHTML = data.identifications.map(i => {
                const confidence = (i.confidence_score * 100).toFixed(1);
                const confClass = confidence >= 90 ? 'high' : confidence >= 70 ? 'medium' : 'low';
                const isNew = i.confidence_score === 1.0 && !i.is_bot;
                return ` + "`" + `
                    <tr>
                        <td>${i.visitor_id.substring(0, 8)}...</td>
                        <td>${i.ip_address}</td>
                        <td class="confidence confidence-${confClass}">${confidence}%</td>
                        <td><span class="badge badge-${isNew ? 'new' : 'returning'}">${isNew ? 'NEW' : 'RETURNING'}</span></td>
                        <td>${new Date(i.created_at).toLocaleString()}</td>
                    </tr>
                ` + "`" + `;
            }).join('');
        }

        loadMetrics();
        loadIdentifications();
        setInterval(() => { loadMetrics(); loadIdentifications(); }, 5000);
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
