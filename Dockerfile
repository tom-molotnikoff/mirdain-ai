# Build stage: compile the mirdain-bridge TypeScript module.
FROM node:20-slim AS bridge-builder
WORKDIR /agent
COPY agent/package.json ./
RUN npm install
COPY agent/ ./
RUN npm run build

# Runtime image.
FROM node:20-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    curl \
    gnupg \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
    | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
    | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update && apt-get install -y --no-install-recommends gh \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled bridge and its runtime dependencies.
COPY --from=bridge-builder /agent/dist /app/dist
COPY --from=bridge-builder /agent/node_modules /app/node_modules

# Directory layout required by the container spec (see #1).
RUN mkdir -p /workspace /run/secrets /skills

WORKDIR /workspace

# Requires MIRDAIN_RUN_ID, MIRDAIN_RUN_SECRET, and MIRDAIN_ORCHESTRATOR_URL.
CMD ["node", "/app/dist/bridge.js"]
