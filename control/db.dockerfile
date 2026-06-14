FROM mysql:8.0

# Schema is managed by versioned migrations (see migrations/ and cmd/migrate).
# MYSQL_DATABASE from docker-compose creates an empty database on first boot.
# Data persists in the mysql_data volume across compose down/restart.
