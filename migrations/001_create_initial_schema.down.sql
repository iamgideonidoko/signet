DROP VIEW IF EXISTS visitor_analytics;

DROP TRIGGER IF EXISTS trigger_update_visitor ON identifications;

DROP FUNCTION IF EXISTS update_visitor_timestamp ();

DROP TABLE IF EXISTS identifications;

DROP TABLE IF EXISTS visitors;

DROP EXTENSION IF EXISTS "uuid-ossp";

