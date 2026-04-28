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

# TODO(#21): install Pi (badlogic/pi-mono) and the mirdain-bridge extension.
# Example (version must be pinned once decided):
#   RUN npm install -g <pi-package>
#   COPY agent/dist/ /app/bridge/

# Directory layout required by the container spec (see #1).
RUN mkdir -p /workspace /run/secrets /skills

WORKDIR /workspace
