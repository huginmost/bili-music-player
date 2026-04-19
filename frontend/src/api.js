const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8765/api'
const MEDIA_BASE = API_BASE.replace(/\/api$/, '')

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {})
    },
    ...options
  })

  const contentType = response.headers.get('content-type') || ''
  const payload = contentType.includes('application/json')
    ? await response.json()
    : await response.text()

  if (!response.ok) {
    const message = typeof payload === 'object' && payload?.error ? payload.error : String(payload)
    throw new Error(message)
  }

  return payload
}

export function fetchLibrary() {
  return request('/library')
}

export function fetchSettings() {
  return request('/settings')
}

export function saveSettings(settings) {
  return request('/settings', {
    method: 'PUT',
    body: JSON.stringify(settings)
  })
}

export function importVideo(id) {
  return request('/library/import/video', {
    method: 'POST',
    body: JSON.stringify({ id })
  })
}

export function importList(id) {
  return request('/library/import/list', {
    method: 'POST',
    body: JSON.stringify({ id })
  })
}

export function refreshTrackAudio(playlistTitle, bvid) {
  return request('/tracks/refresh', {
    method: 'POST',
    body: JSON.stringify({ playlistTitle, bvid })
  })
}

export function prefetchTrackAudio(playlistTitle, bvids) {
  return request('/tracks/prefetch', {
    method: 'POST',
    body: JSON.stringify({ playlistTitle, bvids })
  })
}

export function deletePlaylist(title) {
  return request('/playlists', {
    method: 'DELETE',
    body: JSON.stringify({ title })
  })
}

export function deleteTrack(playlistTitle, bvid) {
  return request('/tracks', {
    method: 'DELETE',
    body: JSON.stringify({ playlistTitle, bvid })
  })
}

export function downloadTrack(playlistTitle, bvid) {
  return request('/downloads', {
    method: 'POST',
    body: JSON.stringify({ playlistTitle, bvid })
  })
}

export function coverURL(src) {
  return `${MEDIA_BASE}/media/cover?src=${encodeURIComponent(src)}`
}

export function audioURL(playlistTitle, bvid) {
  return `${MEDIA_BASE}/media/audio?playlistTitle=${encodeURIComponent(playlistTitle)}&bvid=${encodeURIComponent(bvid)}`
}
