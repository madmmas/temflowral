# Static Module Federation remote for the workflow builder.
# Build context: repository root.
FROM node:20-bookworm AS build
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./frontend/
COPY packages/canvas-remote/package.json ./packages/canvas-remote/
WORKDIR /src/packages/canvas-remote
RUN npm install
WORKDIR /src
COPY frontend ./frontend
COPY packages/canvas-remote ./packages/canvas-remote
WORKDIR /src/packages/canvas-remote
RUN npm run build

FROM nginx:1.27-alpine
COPY packages/canvas-remote/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/packages/canvas-remote/dist /usr/share/nginx/html
EXPOSE 80
