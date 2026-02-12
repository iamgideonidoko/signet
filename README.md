# signet

HiFi Browser Fingerprinting with probabilistic matching that just works. Signet maintains stable visitor IDs across browser updates via weighted Jaccard similarity (≥75% threshold).

## Quick Start

```bash
git clone https://github.com/iamgideonidoko/signet.git && cd signet
cp .env.example .env  # Edit: set DATABASE_URL with your password
make docker-up && make migrate
curl http://localhost:6969/health
```

Dashboard: http://localhost:6969/dashboard

## How It Works

```
Browser (TS Agent) → Go API → Redis Cache → Similarity Engine → PostgreSQL
                                    ↓
                            visitor_id + confidence
```

**Algorithm:**

1. Compute SHA-256 hardware hash (canvas + audio + webgl)
2. Check Redis cache → HIT: return visitor_id | MISS: continue
3. Query DB for candidates in same /24 subnet
4. Calculate weighted Jaccard similarity (≥0.75 threshold)
5. Match found: reuse visitor_id (healed) | No match: create new
6. Cache for 48h, return response

## Use Cases

- **Fraud Detection:** Track users across cookie clearing and incognito mode
- **A/B Testing:** Consistent bucketing without cookies
- **Analytics:** Accurate unique visitors (survives deletion)
- **Paywall/Rate Limiting:** Enforce limits by device fingerprint
- **User Recognition:** Identify returning users without login

## Usage

**Client:**

```html
<script src="https://your-domain.com/agent.js"></script>
<script>
  Signet.identify("https://your-domain.com/v1/identify").then((result) => {
    console.log(result.visitor_id, result.confidence, result.is_new);
  });
</script>
```

**API:**

```bash
POST /v1/identify
{
  "signals": {
    "canvas_2d_hash": "...",
    "audio_hash": "...",
    "webgl_vendor": "...",
    ...
  }
}

# Response
{
  "visitor_id": "uuid",
  "confidence": 0.95,  # ≥0.75 = healed match
  "is_new": false,
  "request_id": "uuid"
}
```

**Endpoints:**

- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /dashboard` - Analytics UI
- `GET /agent.js` - Agent script

## Development

**Stack:** Go 1.25+, Fiber, PostgreSQL 15+, Redis 7+, TypeScript (zero deps)

```bash
make install-deps  # Install dependencies
make build        # Build API + Agent
make test         # Run tests
make dev          # Start dev mode
```

## Contributing

**Priority areas:** Fingerprinting techniques, performance optimization, security audits, ML similarity scoring.

## Production Deployment

```bash
cp .env.example .env  # Set DATABASE_URL, REDIS_URL, CORS_ORIGINS
make build-agent && make docker-up && make migrate
```

**Options:** Docker Compose (recommended) | Kubernetes | AWS ECS | Cloud Run | Bare Metal

## Roadmap

- WebSocket support, enhanced bot detection
- ML similarity scoring, Redis Cluster, geo enrichment
- Multi-region deployment, GraphQL API, fraud detection engine

## License

MIT - See [LICENSE](LICENSE)

## Support

[Issues](https://github.com/iamgideonidoko/signet/issues) | [Discussions](https://github.com/iamgideonidoko/signet/discussions) | [Docs](GETTING_STARTED.md)

Reference papers:

- [ThresholdFP: Enhanced Durability in Browser Fingerprinting](https://research.sabanciuniv.edu/52205/1/ThresholdFP.pdf) by [ELIF ECEM ŞAMLIOĞLU](http://linkedin.com/in/ecem-%C5%9Faml%C4%B1o%C4%9Flu-b688441b5/), [SINAN EMRE TAŞÇI](linkedin.com/in/sinan-emre-tasci?originalSubdomain=tr), [MUHAMMED FATIH GÜLŞEN](https://www.linkedin.com/in/fatihglsn/), and [AND ALBERT LEVI](https://scholar.google.com/citations?user=ls6NkwEAAAAJ&hl=tr)
- [How Unique is Whose Web Browser? The Role of Demographics in Browser Fingerprinting](https://arxiv.org/pdf/2410.06954) by [Alex Berke](https://scholar.google.com/citations?user=1SaHM5UAAAAJ&hl=en), [Enrico Bacis](https://scholar.google.com/citations?user=y9SM26UAAAAJ&hl=en), [Badih Ghazi](https://scholar.google.com/citations?user=GBJLTN8AAAAJ&hl=en), [Pritish Kamath](https://scholar.google.com/citations?user=1JFARhUAAAAJ&hl=en), [Ravi Kumar](https://scholar.google.com/citations?user=J_XhIsgAAAAJ&hl=en), [Robin Lassonde](linkedin.com/in/robin-lassonde-5140154b), [Pasin Manurangsi](https://scholar.google.com/citations?user=35hM-PkAAAAJ&hl=en), and [Umar Syed](https://scholar.google.com/citations?user=zKORw8wAAAAJ&hl=en)
- [Browser Fingerprinting Using WebAssembly](https://arxiv.org/pdf/2506.00719) by [Mordechai Guri](https://scholar.google.com/citations?user=F8gvBUkAAAAJ&hl=en), and Dor Fibert
- [Cascading Spy Sheets: Exploiting the Complexity of Modern CSS for Email and Browser Fingerprinting](https://publications.cispa.de/articles/conference_contribution/Cascading_Spy_Sheets_Exploiting_the_Complexity_of_Modern_CSS_for_Email_and_Browser_Fingerprinting/27194472) by [Leon Trampert](https://cispa.de/en/people/c01letr), [Daniel Weber](https://cispa.de/en/people/daniel.weber), [Lukas Gerlach](https://cispa.de/en/people/c01luge), [Christian Rossow](https://cispa.de/en/people/rossow), and [Michael Schwarz](https://cispa.de/en/people/c02misc)
- [Beyond the Crawl: Unmasking Browser Fingerprinting in Real User Interactions](https://arxiv.org/pdf/2502.01608) by [Meenatchi Sundaram Mutu Selva Annamalai](https://scholar.google.com/citations?user=zYVEyL4AAAAJ&hl=en), [Emiliano De Cristofaro](https://scholar.google.com/citations?user=1wfzUuEAAAAJ&hl=en), and [Igor Bilogrevic](https://scholar.google.com/citations?user=7h8KipcAAAAJ&hl=en)
- [The First Early Evidence of the Use of Browser Fingerprinting for Online Tracking](https://arxiv.org/pdf/2409.15656) by [Zengrui Liu](https://scholar.google.com/citations?user=aUJpMPYAAAAJ&hl=en), [Jimmy Dani](https://scholar.google.com/citations?user=occFbWAAAAAJ&hl=en), [Yinzhi Cao](https://scholar.google.com/citations?user=0jBP_aEAAAAJ&hl=en), [Shujiang Wu](https://scholar.google.com/citations?user=3yQomhcAAAAJ&hl=en), and [Nitesh Saxena](https://scholar.google.com/citations?user=_x5BEjoAAAAJ&hl=en)
- [Mitigating Browser Fingerprinting in Web Specifications](https://www.w3.org/TR/fingerprinting-guidance/) by [W3C](https://www.w3.org/)
- [Fingerprinting and Tracing Shadows: The Development and Impact of Browser Fingerprinting](https://www.thinkmind.org/articles/securware_2024_2_160_30065.pdf) by [Alexander Lawall](https://www.linkedin.com/in/prof-dr-alexander-lawall-b927a131b/)
