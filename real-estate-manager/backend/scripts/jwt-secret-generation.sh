#!/bin/bash

echo "Generating JWT Secret..."

# Generate a random 64-character string
JWT_SECRET=$(openssl rand -hex 32)

echo "Generated JWT Secret: $JWT_SECRET"
echo ""
echo "Add this to your environment:"
echo "JWT_SECRET=$JWT_SECRET"
echo ""
echo "For Docker Compose, update the JWT_SECRET in docker-compose.yml"