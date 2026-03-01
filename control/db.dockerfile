FROM mysql:8.0

# Copy the initialization script
COPY up.sql /docker-entrypoint-initdb.d/

# Ensure the script is readable
RUN chmod 644 /docker-entrypoint-initdb.d/up.sql
