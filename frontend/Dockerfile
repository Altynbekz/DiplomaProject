FROM node:26.13.0@sha256:e7d748ae5e42807dde2a43559ed366ec6ecf8a4ebe79d31c3cb7da4eaf1282db
WORKDIR /app
# Vite writes its dependency cache as the runtime user.
RUN chown node:node /app
USER node
COPY --chown=node:node package.json package-lock.json ./
RUN npm ci
COPY --chown=node:node . .
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]
