package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/iamgideonidoko/signet/internal/config"
	"github.com/iamgideonidoko/signet/internal/models"
	"github.com/iamgideonidoko/signet/internal/repository"
	"github.com/iamgideonidoko/signet/pkg/cache"
	"github.com/iamgideonidoko/signet/pkg/similarity"
)

type IdentificationService struct {
	repo       *repository.Repository
	cache      *cache.Cache
	calculator *similarity.Calculator
	config     *config.FingerprintConfig
}

func NewIdentificationService(
	repo *repository.Repository,
	cache *cache.Cache,
	cfg *config.FingerprintConfig,
) *IdentificationService {
	weights := similarity.Weights{
		Hardware:    cfg.HardwareWeight,
		Environment: cfg.EnvironmentWeight,
		Software:    cfg.SoftwareWeight,
	}

	return &IdentificationService{
		repo:       repo,
		cache:      cache,
		calculator: similarity.NewCalculator(weights),
		config:     cfg,
	}
}

// Identify performs the "Healer" logic: probabilistic matching with self-healing.
func (s *IdentificationService) Identify(ctx context.Context, req models.IdentifyRequest) (*models.IdentifyResponse, error) {
	// Generate hardware hash for cache lookup
	hardwareHash := similarity.ComputeHardwareHash(req.Signals)

	// Step 1: Check Redis cache for quick match
	cachedVisitorID, err := s.cache.GetVisitorID(ctx, hardwareHash)
	if err == nil && cachedVisitorID != "" {
		visitorUUID, _ := uuid.Parse(cachedVisitorID)

		// Create identification record
		ident := &models.Identification{
			RequestID:       uuid.New(),
			VisitorID:       visitorUUID,
			IPAddress:       req.IPAddress,
			Signals:         req.Signals,
			ConfidenceScore: 1.0, // Cache hit = high confidence
			CreatedAt:       time.Now(),
			HardwareHash:    hardwareHash,
			IsBot:           s.detectBot(req.Signals),
		}

		if err := s.repo.CreateIdentification(ctx, ident); err != nil {
			return nil, fmt.Errorf("failed to save identification: %w", err)
		}

		// Track metric
		_ = s.cache.IncrementMetric(ctx, "cache_hits")

		return &models.IdentifyResponse{
			VisitorID:  visitorUUID,
			Confidence: 1.0,
			IsNew:      false,
			RequestID:  ident.RequestID,
		}, nil
	}

	// Step 2: Extract feature vector from incoming signals
	incomingVector := s.calculator.ExtractFeatures(req.Signals)

	// Step 3: Get IP subnet for similarity search
	ipSubnet := s.extractIPSubnet(req.IPAddress)

	// Step 4: Find similar visitors in the same subnet
	candidates, err := s.repo.FindSimilarVisitors(ctx, ipSubnet, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar visitors: %w", err)
	}

	// Step 5: Compute similarity scores and find best match
	var bestMatch *models.Identification
	var bestScore = 0.0

	for _, candidate := range candidates {
		candidateVector := s.calculator.ExtractFeatures(candidate.Signals)
		score := s.calculator.JaccardSimilarity(incomingVector, candidateVector)

		if score > bestScore {
			bestScore = score
			bestMatch = &candidate
		}
	}

	// Step 6: Decide if we found a match or create new visitor
	var visitorID uuid.UUID
	var confidence float64
	var isNew bool

	if bestScore >= s.config.SimilarityThreshold && bestMatch != nil {
		// Match found! Use existing visitorID (Self-Healing)
		visitorID = bestMatch.VisitorID
		confidence = bestScore
		isNew = false

		// Update cache with this hardware hash mapping
		_ = s.cache.SetVisitorID(ctx, hardwareHash, visitorID.String())
		_ = s.cache.IncrementMetric(ctx, "healed_identifications")
	} else {
		// No match - create new visitor
		visitor, err := s.repo.CreateVisitor(ctx, req.IPAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to create visitor: %w", err)
		}

		visitorID = visitor.VisitorID
		confidence = 1.0 // New visitor = perfect confidence
		isNew = true

		// Cache the new visitor
		_ = s.cache.SetVisitorID(ctx, hardwareHash, visitorID.String())
		_ = s.cache.IncrementMetric(ctx, "new_visitors")
	}

	// Step 7: Create identification record
	ident := &models.Identification{
		RequestID:       uuid.New(),
		VisitorID:       visitorID,
		IPAddress:       req.IPAddress,
		Signals:         req.Signals,
		ConfidenceScore: confidence,
		CreatedAt:       time.Now(),
		HardwareHash:    hardwareHash,
		IsBot:           s.detectBot(req.Signals),
	}

	if err := s.repo.CreateIdentification(ctx, ident); err != nil {
		return nil, fmt.Errorf("failed to save identification: %w", err)
	}

	return &models.IdentifyResponse{
		VisitorID:  visitorID,
		Confidence: confidence,
		IsNew:      isNew,
		RequestID:  ident.RequestID,
	}, nil
}

// extractIPSubnet returns the /24 subnet of an IP address.
func (s *IdentificationService) extractIPSubnet(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return ip // Return as-is if not IPv4
	}
	return fmt.Sprintf("%s.%s.%s.0/24", parts[0], parts[1], parts[2])
}

// detectBot checks for bot/headless browser indicators.
func (s *IdentificationService) detectBot(signals models.Signals) bool {
	// Check comprehensive bot signals
	if signals.WebDriver || signals.PhantomPresent ||
		signals.SeleniumPresent || signals.AutomationPresent {
		return true
	}

	// Headless Chrome
	if signals.HeadlessChrome {
		return true
	}

	// Missing critical hardware signals (common in headless)
	if signals.Canvas2DHash == "" || signals.Canvas2DHash == "error" ||
		signals.AudioHash == "" || signals.AudioHash == "error" {
		return true
	}

	// Impossible hardware combinations
	if signals.HardwareConcurrency == 0 || signals.DeviceMemory == 0 {
		return true
	}

	// Suspicious WebGL
	if signals.WebGLVendor == "Brian Paul" || // Mesa software renderer
		signals.WebGLRenderer == "Google SwiftShader" {
		return true
	}

	return false
}

// GetAnalytics retrieves visitor analytics.
func (s *IdentificationService) GetAnalytics(ctx context.Context, days int) ([]models.VisitorAnalytics, error) {
	return s.repo.GetAnalytics(ctx, days)
}

// GetRecentIdentifications retrieves recent identifications with pagination.
func (s *IdentificationService) GetRecentIdentifications(ctx context.Context, limit, offset int) ([]models.Identification, error) {
	return s.repo.GetRecentIdentifications(ctx, limit, offset)
}
