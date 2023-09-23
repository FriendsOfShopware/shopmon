FROM mcr.microsoft.com/devcontainers/javascript-node:1-20-bullseye

RUN sudo corepack enable --install-directory /usr/local/bin && \
    echo 'alias yarn="corepack yarn"' >> /home/node/.bash_aliases && \
    echo 'alias yarnpkg="corepack yarnpkg"' >> /home/node/.bash_aliases && \
    echo 'alias pnpm="corepack pnpm"' >> /home/node/.bash_aliases && \
    echo 'alias pnpx="corepack pnpx"' >> /home/node/.bash_aliases && \
    echo 'alias npm="corepack npm"' >> /home/node/.bash_aliases && \
    echo 'alias npx="corepack npx"' >> /home/node/.bash_aliases
