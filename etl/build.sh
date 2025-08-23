#!/bin/bash

set -e

echo "Building ETL..."
cd "$(dirname "$0")"

# Build the application
go build -o bin/etl ./cmd/etl

echo "Build complete! Binary available at bin/etl"
echo ""
echo "Usage examples:"
echo "  ./bin/etl                           # Full import"
echo "  ./bin/etl -after today              # Import today's changes"
echo "  ./bin/etl -after this-week          # Import this week's changes"
echo "  ./bin/etl -after 2024-01-01T00:00:00Z # Import since specific date"
echo ""
echo "Available date keywords:"
echo "  - today, yesterday"
echo "  - this-week, last-week"
echo "  - this-month, last-month"
echo "  - Or use RFC3339 format: 2024-01-01T00:00:00Z"
