FROM sitespeedio/sitespeed.io:latest

# Create app directory
WORKDIR /app

# Copy application code
COPY . .

RUN npm install

# Create results directory with proper permissions
RUN mkdir -p /app/results && chmod 755 /app/results

# Expose port
EXPOSE 3001

# Override entrypoint and use absolute path to node
ENTRYPOINT []
CMD ["/usr/local/bin/node", "--no-warnings=ExperimentalWarning", "--experimental-strip-types", "index.ts"]
