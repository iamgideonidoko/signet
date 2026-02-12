-- Description: Create visitors and identifications tables
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Visitors table: Stores unique visitor identities
CREATE TABLE visitors (
  visitor_id uuid PRIMARY KEY DEFAULT uuid_generate_v4 (),
  created_at timestamp NOT NULL DEFAULT NOW(),
  updated_at timestamp NOT NULL DEFAULT NOW(),
  trust_score float NOT NULL DEFAULT 1.0,
  first_seen_ip inet,
  last_seen_ip inet,
  visit_count integer NOT NULL DEFAULT 1
);

CREATE INDEX idx_visitors_created_at ON visitors (created_at);

CREATE INDEX idx_visitors_trust_score ON visitors (trust_score);

-- Identifications table: Stores all fingerprint submissions
CREATE TABLE identifications (
  request_id uuid PRIMARY KEY DEFAULT uuid_generate_v4 (),
  visitor_id uuid NOT NULL REFERENCES visitors (visitor_id) ON DELETE CASCADE,
  ip_address inet NOT NULL,
  ip_subnet cidr GENERATED ALWAYS AS (host(ip_address)::inet & '255.255.255.0'::inet) STORED,
  user_agent text,
  signals jsonb NOT NULL,
  confidence_score float NOT NULL,
  created_at timestamp NOT NULL DEFAULT NOW(),
  hardware_hash text NOT NULL,
  is_bot boolean DEFAULT FALSE
);

CREATE INDEX idx_identifications_visitor_id ON identifications (visitor_id);

CREATE INDEX idx_identifications_created_at ON identifications (created_at);

CREATE INDEX idx_identifications_hardware_hash ON identifications (hardware_hash);

CREATE INDEX idx_identifications_ip_subnet ON identifications (ip_subnet);

CREATE INDEX idx_identifications_signals ON identifications USING GIN (signals);

-- Trigger to update visitors.updated_at
CREATE OR REPLACE FUNCTION update_visitor_timestamp ()
  RETURNS TRIGGER
  AS $$
BEGIN
  UPDATE
    visitors
  SET
    updated_at = NOW(),
    last_seen_ip = NEW.ip_address,
    visit_count = visit_count + 1
  WHERE
    visitor_id = NEW.visitor_id;
  RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_visitor
  AFTER INSERT ON identifications
  FOR EACH ROW
  EXECUTE FUNCTION update_visitor_timestamp ();

-- Analytics view for dashboard
CREATE VIEW visitor_analytics AS
SELECT
  DATE(i.created_at) AS date,
  COUNT(DISTINCT i.visitor_id) AS unique_visitors,
  COUNT(*) AS total_requests,
  AVG(i.confidence_score) AS avg_confidence,
  SUM(
    CASE WHEN i.is_bot THEN
      1
    ELSE
      0
    END) AS bot_requests
FROM
  identifications i
GROUP BY
  DATE(i.created_at);

