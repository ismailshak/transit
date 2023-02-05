# See here for image contents: https://github.com/devcontainers/images/blob/v0.2.40/src/go/manifest.json
ARG VARIANT="1"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

ARG INSTALL_NODE="true"
ARG NODE_VERSION="lts/*"

# Install node LTS (to get npm)
RUN if [ "${INSTALL_NODE}" = "true" ]; then su vscode -c "source /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

# Uncomment this section to install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && apt-get -y install --no-install-recommends <your-package-list-here>

# Install the last released version of transit
RUN go install github.com/ismailshak/transit@latest

# Installing prettier for markdown formatting
RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g prettier" 2>&1