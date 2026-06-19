#!/usr/bin/env bash
set -euo pipefail

export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

LOCK_FILE=${LOCK_FILE:-/var/lock/tokenhut-deploy.lock}
LOG_FILE=${LOG_FILE:-/var/log/tokenhut-deploy.log}
DEPLOY_DIR=${DEPLOY_DIR:-/opt/tokenhut}
REPO_URL=${REPO_URL:-https://github.com/haibin-zhou/sub2api.git}
BRANCH=${BRANCH:-main}
IMAGE_REPO=${IMAGE_REPO:-ghcr.io/haibin-zhou/sub2api}
COMPOSE_FILE=${COMPOSE_FILE:-docker-compose.local.yml}
FORCE=false

if [ "${1:-}" = "--force" ]; then
  FORCE=true
fi

exec 9>"${LOCK_FILE}"
if ! flock -n 9; then
  echo "$(date -Is) another deployment is already running"
  exit 0
fi

log() {
  echo "$(date -Is) $*" | tee -a "${LOG_FILE}"
}

log "checking ${REPO_URL} ${BRANCH}"
remote_sha="$(git ls-remote --heads "${REPO_URL}" "${BRANCH}" | cut -f1)"
if [ -z "${remote_sha}" ]; then
  log "ERROR: branch ${BRANCH} not found"
  exit 1
fi

short_sha="${remote_sha:0:12}"
image="${IMAGE_REPO}:sha-${short_sha}"

deployed_sha="$(cat "${DEPLOY_DIR}/.deployed_commit" 2>/dev/null || true)"
if [ "${FORCE}" != "true" ] && [ "${remote_sha}" = "${deployed_sha}" ]; then
  log "no changes: ${remote_sha} already deployed"
  exit 0
fi

log "pulling image ${image}"
for attempt in $(seq 1 60); do
  if docker pull "${image}"; then
    break
  fi
  if [ "${attempt}" = "60" ]; then
    log "ERROR: image ${image} was not available after waiting"
    exit 1
  fi
  log "image not ready yet; retrying in 10s (${attempt}/60)"
  sleep 10
done

log "updating ${DEPLOY_DIR}/.env image setting"
if grep -q '^TOKENHUT_IMAGE=' "${DEPLOY_DIR}/.env"; then
  sed -i "s#^TOKENHUT_IMAGE=.*#TOKENHUT_IMAGE=${image}#" "${DEPLOY_DIR}/.env"
else
  printf '\nTOKENHUT_IMAGE=%s\n' "${image}" >> "${DEPLOY_DIR}/.env"
fi

log "recreating application container"
cd "${DEPLOY_DIR}"
docker compose -f "${COMPOSE_FILE}" up -d --no-deps --force-recreate sub2api

log "waiting for health check"
for _ in $(seq 1 60); do
  if curl -fsS http://127.0.0.1:8080/health >/dev/null; then
    echo "${remote_sha}" > "${DEPLOY_DIR}/.deployed_commit"
    log "deployed ${remote_sha} successfully"
    exit 0
  fi
  sleep 2
done

log "ERROR: health check failed after deployment"
docker compose -f "${COMPOSE_FILE}" logs --tail=120 sub2api | tee -a "${LOG_FILE}"
exit 1
