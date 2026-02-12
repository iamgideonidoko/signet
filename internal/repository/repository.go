package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/iamgideonidoko/signet/internal/models"
	"github.com/iamgideonidoko/signet/pkg/logger"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(dsn string, maxConns, maxIdleConns int) (*Repository, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	return &Repository{db: db}, nil
}

// CreateVisitor creates a new visitor record.
func (r *Repository) CreateVisitor(ctx context.Context, ipAddress string) (*models.Visitor, error) {
	visitor := &models.Visitor{
		VisitorID:   uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		TrustScore:  1.0,
		FirstSeenIP: &ipAddress,
		LastSeenIP:  &ipAddress,
		VisitCount:  1,
	}

	query := `
		INSERT INTO visitors (visitor_id, created_at, updated_at, trust_score, first_seen_ip, last_seen_ip, visit_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		visitor.VisitorID, visitor.CreatedAt, visitor.UpdatedAt,
		visitor.TrustScore, visitor.FirstSeenIP, visitor.LastSeenIP, visitor.VisitCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create visitor: %w", err)
	}

	return visitor, nil
}

// GetVisitor retrieves a visitor by ID.
func (r *Repository) GetVisitor(ctx context.Context, visitorID uuid.UUID) (*models.Visitor, error) {
	var visitor models.Visitor
	query := `SELECT * FROM visitors WHERE visitor_id = $1`

	if err := r.db.GetContext(ctx, &visitor, query, visitorID); err != nil {
		return nil, fmt.Errorf("failed to get visitor: %w", err)
	}

	return &visitor, nil
}

// CreateIdentification stores a new fingerprint identification.
func (r *Repository) CreateIdentification(ctx context.Context, ident *models.Identification) error {
	signalsJSON, err := json.Marshal(ident.Signals)
	if err != nil {
		return fmt.Errorf("failed to marshal signals: %w", err)
	}

	query := `
		INSERT INTO identifications 
		(request_id, visitor_id, ip_address, user_agent, signals, confidence_score, created_at, hardware_hash, is_bot)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.ExecContext(ctx, query,
		ident.RequestID, ident.VisitorID, ident.IPAddress, ident.UserAgent,
		signalsJSON, ident.ConfidenceScore, ident.CreatedAt, ident.HardwareHash, ident.IsBot,
	)
	if err != nil {
		return fmt.Errorf("failed to create identification: %w", err)
	}

	return nil
}

// FindSimilarVisitors finds visitors with similar fingerprints in the same IP subnet.
func (r *Repository) FindSimilarVisitors(ctx context.Context, ipSubnet string, limit int) ([]models.Identification, error) {
	query := `
		SELECT DISTINCT ON (visitor_id) *
		FROM identifications
		WHERE ip_subnet = $1::cidr
		ORDER BY visitor_id, created_at DESC
		LIMIT $2
	`

	var identifications []models.Identification
	rows, err := r.db.QueryxContext(ctx, query, ipSubnet, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar visitors: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Warn("Failed to close database rows", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	for rows.Next() {
		var ident models.Identification
		var signalsJSON []byte

		err := rows.Scan(
			&ident.RequestID, &ident.VisitorID, &ident.IPAddress, &ident.UserAgent,
			&signalsJSON, &ident.ConfidenceScore, &ident.CreatedAt, &ident.HardwareHash, &ident.IsBot,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan identification: %w", err)
		}

		if err := json.Unmarshal(signalsJSON, &ident.Signals); err != nil {
			return nil, fmt.Errorf("failed to unmarshal signals: %w", err)
		}

		identifications = append(identifications, ident)
	}

	return identifications, nil
}

// UpdateVisitorSignals updates a visitor's signals with new data (self-healing).
func (r *Repository) UpdateVisitorSignals(ctx context.Context, visitorID uuid.UUID, newSignals models.Signals) error {
	// This could merge new signals with existing ones
	// For now, we just create a new identification record
	return nil
}

// GetAnalytics retrieves visitor analytics for the dashboard.
func (r *Repository) GetAnalytics(ctx context.Context, days int) ([]models.VisitorAnalytics, error) {
	query := `
		SELECT 
			date,
			unique_visitors,
			total_requests,
			avg_confidence,
			bot_requests
		FROM visitor_analytics
		WHERE date >= CURRENT_DATE - $1::integer
		ORDER BY date DESC
	`

	var analytics []models.VisitorAnalytics
	if err := r.db.SelectContext(ctx, &analytics, query, days); err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return analytics, nil
}

// GetRecentIdentifications retrieves recent identifications with pagination.
func (r *Repository) GetRecentIdentifications(ctx context.Context, limit, offset int) ([]models.Identification, error) {
	query := `
		SELECT * FROM identifications
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var identifications []models.Identification
	rows, err := r.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent identifications: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Warn("Failed to close database rows", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	for rows.Next() {
		var ident models.Identification
		var signalsJSON []byte

		err := rows.Scan(
			&ident.RequestID, &ident.VisitorID, &ident.IPAddress, &ident.UserAgent,
			&signalsJSON, &ident.ConfidenceScore, &ident.CreatedAt, &ident.HardwareHash, &ident.IsBot,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan identification: %w", err)
		}

		if err := json.Unmarshal(signalsJSON, &ident.Signals); err != nil {
			return nil, fmt.Errorf("failed to unmarshal signals: %w", err)
		}

		identifications = append(identifications, ident)
	}

	return identifications, nil
}

// Close closes the database connection.
func (r *Repository) Close() error {
	return r.db.Close()
}

// BeginTx starts a transaction.
func (r *Repository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}
