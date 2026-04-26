<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  audioURL,
  coverURL,
  deletePlaylist,
  deleteTrack,
  downloadTrack,
  fetchLibrary,
  fetchSettings,
  importList,
  importVideo,
  prefetchTrackAudio,
  refreshTrackAudio,
  saveSettings
} from './api'

const PREFETCH_SIZE = 3

const library = ref({})
const loading = ref(true)
const error = ref('')
const statusText = ref('夜色已就绪。')
const activePlaylistTitle = ref('')
const activeTrackIndex = ref(0)
const playMode = ref('sequence')
const search = ref('')
const audioRef = ref(null)
const isPlaying = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const volume = ref(0.72)
const importVideoID = ref('')
const importListID = ref('')
const importBusy = ref(false)
const shuffleQueue = ref([])
const historyStack = ref([])
const settingsReady = ref(false)
const pendingResumeTime = ref(null)
const playbackRetryCount = ref(0)
let restoreSettings = null
let saveTimer = null
let prefetchKey = ''
let playRequestId = 0

const playlists = computed(() =>
  Object.entries(library.value).map(([title, tracks]) => ({
    title,
    tracks
  }))
)

const filteredPlaylists = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  if (!keyword) {
    return playlists.value
  }

  return playlists.value
    .map((playlist) => ({
      ...playlist,
      tracks: playlist.tracks.filter((track) =>
        [track.title, track.bvid].some((value) => String(value || '').toLowerCase().includes(keyword))
      )
    }))
    .filter((playlist) => playlist.title.toLowerCase().includes(keyword) || playlist.tracks.length > 0)
})

const activePlaylist = computed(() =>
  playlists.value.find((playlist) => playlist.title === activePlaylistTitle.value) || null
)

const visiblePlaylist = computed(() =>
  filteredPlaylists.value.find((playlist) => playlist.title === activePlaylistTitle.value) ||
  filteredPlaylists.value[0] ||
  null
)

const activeTrack = computed(() => {
  const playlist = activePlaylist.value
  if (!playlist || !playlist.tracks.length) {
    return null
  }
  return playlist.tracks[activeTrackIndex.value] || playlist.tracks[0]
})

const upcomingTracks = computed(() => {
  const playlist = activePlaylist.value
  if (!playlist || !playlist.tracks.length) {
    return []
  }

  if (playMode.value === 'shuffle') {
    return shuffleQueue.value
      .slice(0, PREFETCH_SIZE)
      .map((index) => playlist.tracks[index])
      .filter(Boolean)
  }

  const result = []
  for (let step = 1; step <= PREFETCH_SIZE; step += 1) {
    const index = activeTrackIndex.value + step
    if (index >= playlist.tracks.length) {
      break
    }
    result.push(playlist.tracks[index])
  }
  return result
})

function selectPlaylist(title) {
  activePlaylistTitle.value = title
}

function resolveTrackIndexByBVID(bvid) {
  const playlist = activePlaylist.value
  if (!playlist) {
    return -1
  }
  return playlist.tracks.findIndex((track) => track.bvid === bvid)
}

function syncVolume() {
  if (audioRef.value) {
    audioRef.value.volume = volume.value
  }
}

function isInterruptedPlayRequest(err) {
  const message = String(err?.message || err || '')
  return (
    err?.name === 'AbortError' ||
    message.includes('The play() request was interrupted') ||
    message.includes('interrupted by a new load request')
  )
}

function shuffleIndices(indices) {
  const result = [...indices]
  for (let i = result.length - 1; i > 0; i -= 1) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[result[i], result[j]] = [result[j], result[i]]
  }
  return result
}

function rebuildShuffleQueue(currentIndex = activeTrackIndex.value) {
  const playlist = activePlaylist.value
  if (!playlist || !playlist.tracks.length) {
    shuffleQueue.value = []
    historyStack.value = []
    return
  }

  const remaining = playlist.tracks
    .map((_, index) => index)
    .filter((index) => index !== currentIndex)

  shuffleQueue.value = shuffleIndices(remaining)
}

function resetShuffleState(currentIndex = activeTrackIndex.value) {
  historyStack.value = []
  rebuildShuffleQueue(currentIndex)
}

function pushHistory(index) {
  if (index < 0) {
    return
  }
  const last = historyStack.value[historyStack.value.length - 1]
  if (last !== index) {
    historyStack.value = [...historyStack.value, index]
  }
}

function takeNextShuffleIndex() {
  const playlist = activePlaylist.value
  if (!playlist || !playlist.tracks.length) {
    return -1
  }

  if (!shuffleQueue.value.length) {
    rebuildShuffleQueue(activeTrackIndex.value)
  }

  const nextIndex = shuffleQueue.value[0]
  shuffleQueue.value = shuffleQueue.value.slice(1)
  return typeof nextIndex === 'number' ? nextIndex : -1
}

function scheduleSaveSettings() {
  if (!settingsReady.value) {
    return
  }

  if (saveTimer) {
    clearTimeout(saveTimer)
  }

  saveTimer = setTimeout(async () => {
    try {
      await saveSettings({
        activePlaylistTitle: activePlaylistTitle.value,
        activeTrackBvid: activeTrack.value?.bvid || '',
        currentTime: currentTime.value,
        playMode: playMode.value,
        volume: volume.value,
        shuffleQueue: shuffleQueue.value,
        historyStack: historyStack.value
      })
    } catch (err) {
      error.value = err.message
    }
  }, 200)
}

function applyRestoredSettings() {
  const settings = restoreSettings
  if (!settings) {
    settingsReady.value = true
    return
  }

  if (settings.activePlaylistTitle && library.value[settings.activePlaylistTitle]) {
    activePlaylistTitle.value = settings.activePlaylistTitle
  }

  if (!activePlaylistTitle.value) {
    activePlaylistTitle.value = Object.keys(library.value)[0] || ''
  }

  const playlist = library.value[activePlaylistTitle.value] || []
  if (settings.activeTrackBvid) {
    const index = playlist.findIndex((track) => track.bvid === settings.activeTrackBvid)
    activeTrackIndex.value = index >= 0 ? index : 0
  }

  currentTime.value = Number.isFinite(settings.currentTime) && settings.currentTime > 0 ? settings.currentTime : 0
  playMode.value = settings.playMode === 'shuffle' ? 'shuffle' : 'sequence'
  volume.value = typeof settings.volume === 'number' ? settings.volume : 0.72
  shuffleQueue.value = Array.isArray(settings.shuffleQueue) ? settings.shuffleQueue : []
  historyStack.value = Array.isArray(settings.historyStack) ? settings.historyStack : []

  if (playMode.value === 'shuffle') {
    const validIndices = new Set(playlist.map((_, index) => index))
    shuffleQueue.value = shuffleQueue.value.filter((index) => validIndices.has(index) && index !== activeTrackIndex.value)
    historyStack.value = historyStack.value.filter((index) => validIndices.has(index))
    if (!shuffleQueue.value.length) {
      rebuildShuffleQueue(activeTrackIndex.value)
    }
  } else {
    shuffleQueue.value = []
    historyStack.value = []
  }

  syncVolume()
  settingsReady.value = true
}

async function loadLibrary(options = {}) {
  const { silent = false } = options
  if (!silent) {
    loading.value = true
  }
  error.value = ''

  try {
    const payload = await fetchLibrary()
    library.value = payload

    const titles = Object.keys(payload)
    if (!titles.length) {
      activePlaylistTitle.value = ''
      activeTrackIndex.value = 0
      shuffleQueue.value = []
      historyStack.value = []
      settingsReady.value = true
      return
    }

    if (!settingsReady.value && restoreSettings) {
      applyRestoredSettings()
    } else if (!payload[activePlaylistTitle.value]) {
      activePlaylistTitle.value = titles[0]
      activeTrackIndex.value = 0
      resetShuffleState(0)
    } else {
      const trackCount = payload[activePlaylistTitle.value]?.length || 0
      if (activeTrackIndex.value >= trackCount) {
        activeTrackIndex.value = 0
      }
      if (playMode.value === 'shuffle' && !shuffleQueue.value.length) {
        rebuildShuffleQueue(activeTrackIndex.value)
      }
    }
  } catch (err) {
    error.value = err.message
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

async function prepareTrack(track = activeTrack.value) {
  if (!track || !activePlaylistTitle.value) {
    return null
  }

  if (track.audio) {
    return track
  }

  statusText.value = `正在更新 ${track.title} 的音频链接...`
  const refreshed = await refreshTrackAudio(activePlaylistTitle.value, track.bvid)
  await loadLibrary({ silent: true })
  statusText.value = `音频已更新：${refreshed.title}`
  return refreshed
}

async function playTrack(index, options = {}) {
  const { preserveShuffle = false, startTime = 0, forceReload = false } = options
  const requestId = ++playRequestId
  const playlist = activePlaylist.value
  if (!playlist || !playlist.tracks[index]) {
    return
  }

  activeTrackIndex.value = index
  error.value = ''

  if (playMode.value === 'shuffle' && !preserveShuffle) {
    resetShuffleState(index)
  }

  try {
    const track = await prepareTrack(playlist.tracks[index])
    if (requestId !== playRequestId) {
      return
    }
    if (!track || !audioRef.value) {
      return
    }

    currentTime.value = Number.isFinite(startTime) && startTime > 0 ? startTime : 0
    pendingResumeTime.value = currentTime.value
    if (!forceReload) {
      playbackRetryCount.value = 0
    }

    const nextAudioSource = audioURL(activePlaylistTitle.value, track.bvid)
    if (forceReload || audioRef.value.src !== nextAudioSource) {
      audioRef.value.src = nextAudioSource
    }
    syncVolume()
    await audioRef.value.play()
    if (requestId !== playRequestId) {
      return
    }
    isPlaying.value = true
    statusText.value = `正在播放：${track.title}`
  } catch (err) {
    if (isInterruptedPlayRequest(err) || requestId !== playRequestId) {
      return
    }
    error.value = err.message
  }
}

async function playTrackByBVID(bvid) {
  const index = resolveTrackIndexByBVID(bvid)
  if (index === -1) {
    return
  }
  await playTrack(index)
}

async function togglePlayback() {
  if (!audioRef.value) {
    return
  }

  if (!activeTrack.value) {
    await playTrack(0)
    return
  }

  if (audioRef.value.paused) {
    try {
      const expectedAudioSource = audioURL(activePlaylistTitle.value, activeTrack.value.bvid)
      if (!audioRef.value.src || audioRef.value.src !== expectedAudioSource) {
        await playTrack(activeTrackIndex.value, { startTime: currentTime.value })
        return
      }

      syncVolume()
      await audioRef.value.play()
      isPlaying.value = true
      statusText.value = `正在播放：${activeTrack.value.title}`
    } catch (err) {
      if (isInterruptedPlayRequest(err)) {
        return
      }
      error.value = err.message
    }
    return
  }

  playRequestId += 1
  audioRef.value.pause()
  isPlaying.value = false
  statusText.value = '已暂停播放。'
  scheduleSaveSettings()
}

async function playNext() {
  const playlist = activePlaylist.value
  if (!playlist || !playlist.tracks.length) {
    return
  }

  if (playMode.value === 'shuffle') {
    const currentIndex = activeTrackIndex.value
    const nextIndex = takeNextShuffleIndex()
    if (nextIndex === -1) {
      statusText.value = '随机队列已刷新。'
      return
    }

    pushHistory(currentIndex)
    await playTrack(nextIndex, { preserveShuffle: true })
    return
  }

  const nextIndex = activeTrackIndex.value + 1
  if (nextIndex >= playlist.tracks.length) {
    isPlaying.value = false
    statusText.value = '队列已经播放完。'
    return
  }

  await playTrack(nextIndex)
}

async function playPrevious() {
  const playlist = activePlaylist.value
  if (!playlist || !playlist.tracks.length) {
    return
  }

  if (playMode.value === 'shuffle') {
    const history = [...historyStack.value]
    const previousIndex = history.pop()
    if (typeof previousIndex !== 'number') {
      return
    }

    shuffleQueue.value = [activeTrackIndex.value, ...shuffleQueue.value]
    historyStack.value = history
    await playTrack(previousIndex, { preserveShuffle: true })
    return
  }

  const previousIndex = Math.max(activeTrackIndex.value - 1, 0)
  await playTrack(previousIndex)
}

async function queuePrefetch() {
  if (!activePlaylistTitle.value || !upcomingTracks.value.length) {
    return
  }

  const missing = upcomingTracks.value
    .filter((track) => track && !track.audio)
    .map((track) => track.bvid)

  if (!missing.length) {
    return
  }

  const nextPrefetchKey = `${activePlaylistTitle.value}:${missing.join(',')}`
  if (prefetchKey === nextPrefetchKey) {
    return
  }
  prefetchKey = nextPrefetchKey

  try {
    await prefetchTrackAudio(activePlaylistTitle.value, missing)
    await loadLibrary({ silent: true })
  } catch (err) {
    error.value = err.message
  } finally {
    if (prefetchKey === nextPrefetchKey) {
      prefetchKey = ''
    }
  }
}

async function importFromVideo() {
  if (!importVideoID.value.trim()) {
    error.value = '请输入 BV 号。'
    return
  }

  importBusy.value = true
  error.value = ''
  try {
    await importVideo(importVideoID.value.trim())
    importVideoID.value = ''
    statusText.value = '视频合集已导入。'
    await loadLibrary()
  } catch (err) {
    error.value = err.message
  } finally {
    importBusy.value = false
  }
}

async function importFromList() {
  if (!importListID.value.trim()) {
    error.value = '请输入收藏夹 ID。'
    return
  }

  importBusy.value = true
  error.value = ''
  try {
    await importList(importListID.value.trim())
    importListID.value = ''
    statusText.value = '收藏夹已导入。'
    await loadLibrary()
  } catch (err) {
    error.value = err.message
  } finally {
    importBusy.value = false
  }
}

async function removePlaylist(title) {
  try {
    await deletePlaylist(title)
    statusText.value = `已删除歌单：${title}`
    await loadLibrary()
  } catch (err) {
    error.value = err.message
  }
}

async function removeTrack(track) {
  try {
    await deleteTrack(activePlaylistTitle.value, track.bvid)
    statusText.value = `已删除歌曲：${track.title}`
    await loadLibrary()
  } catch (err) {
    error.value = err.message
  }
}

async function downloadCurrentTrack(track) {
  try {
    const result = await downloadTrack(activePlaylistTitle.value, track.bvid)
    statusText.value = `已下载到 ${result.filePath}`
    await loadLibrary()
  } catch (err) {
    error.value = err.message
  }
}

function formatTime(value) {
  if (!Number.isFinite(value)) {
    return '00:00'
  }

  const minutes = Math.floor(value / 60)
  const seconds = Math.floor(value % 60)
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
}

function onTimeUpdate() {
  if (!audioRef.value) {
    return
  }

  currentTime.value = audioRef.value.currentTime
  duration.value = audioRef.value.duration || 0
  scheduleSaveSettings()
}

function onSeek(event) {
  if (!audioRef.value) {
    return
  }

  const nextTime = Number(event.target.value)
  audioRef.value.currentTime = nextTime
  currentTime.value = nextTime
  scheduleSaveSettings()
}

function onVolumeInput(event) {
  volume.value = Number(event.target.value)
  syncVolume()
}

function onAudioEnded() {
  playNext()
}

function onLoadedMetadata() {
  if (!audioRef.value) {
    return
  }

  duration.value = audioRef.value.duration || 0
  if (pendingResumeTime.value !== null) {
    audioRef.value.currentTime = pendingResumeTime.value
    currentTime.value = pendingResumeTime.value
    pendingResumeTime.value = null
  }
}

function onAudioPlay() {
  isPlaying.value = true
  scheduleSaveSettings()
}

function onAudioPause() {
  isPlaying.value = false
  scheduleSaveSettings()
}

async function retryCurrentTrack(message) {
  if (!activeTrack.value || playbackRetryCount.value >= 1) {
    error.value = message
    return
  }

  playbackRetryCount.value += 1
  statusText.value = '音频连接异常，正在重试。'

  try {
    await playTrack(activeTrackIndex.value, {
      preserveShuffle: true,
      startTime: currentTime.value,
      forceReload: true
    })
  } catch (err) {
    error.value = err.message
  }
}

function onAudioError() {
  retryCurrentTrack('音频播放失败')
}

function onKeydown(event) {
  if (event.code === 'Space' && event.target.tagName !== 'INPUT') {
    event.preventDefault()
    togglePlayback()
  }
}

watch(activePlaylistTitle, async () => {
  activeTrackIndex.value = 0
  currentTime.value = 0
  duration.value = 0
  pendingResumeTime.value = null
  if (playMode.value === 'shuffle') {
    resetShuffleState(0)
  }
  await queuePrefetch()
  scheduleSaveSettings()
})

watch(playMode, async (mode) => {
  if (mode === 'shuffle') {
    resetShuffleState(activeTrackIndex.value)
    statusText.value = '随机队列已生成。'
  } else {
    historyStack.value = []
    shuffleQueue.value = []
    statusText.value = '已切换到顺序播放。'
  }
  await queuePrefetch()
  scheduleSaveSettings()
})

watch(activeTrackIndex, async () => {
  await queuePrefetch()
  scheduleSaveSettings()
})

watch(volume, () => {
  scheduleSaveSettings()
})

watch(shuffleQueue, () => {
  scheduleSaveSettings()
}, { deep: true })

watch(historyStack, () => {
  scheduleSaveSettings()
}, { deep: true })

onMounted(async () => {
  window.addEventListener('keydown', onKeydown)
  try {
    restoreSettings = await fetchSettings()
  } catch (err) {
    error.value = err.message
  }
  await loadLibrary()
  if (!settingsReady.value) {
    applyRestoredSettings()
  }
  syncVolume()
  await queuePrefetch()
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
  if (saveTimer) {
    clearTimeout(saveTimer)
  }
})
</script>

<template>
  <div class="app-shell">
    <div class="glow glow-left"></div>
    <div class="glow glow-right"></div>

    <aside class="playlist-pane">
      <div class="brand">
        <p class="eyebrow">BMplayer</p>
        <h1>Moonlit Library</h1>
      </div>

      <div class="import-panel">
        <p class="eyebrow">Import</p>
        <label class="search compact">
          <span>视频 BV</span>
          <input v-model="importVideoID" type="text" placeholder="例如 BV1oU1jBXEN8" />
        </label>
        <button class="control accent import-action" :disabled="importBusy" @click="importFromVideo">
          导入视频合集
        </button>

        <label class="search compact">
          <span>收藏夹 ID</span>
          <input v-model="importListID" type="text" placeholder="例如 ml3888553754" />
        </label>
        <button class="control import-action" :disabled="importBusy" @click="importFromList">
          导入收藏夹
        </button>
      </div>

      <label class="search">
        <span>检索歌单 / BV</span>
        <input v-model="search" type="text" placeholder="输入标题或 BV 号" />
      </label>

      <div class="playlist-list">
        <button
          v-for="playlist in filteredPlaylists"
          :key="playlist.title"
          class="playlist-card"
          :class="{ active: playlist.title === activePlaylistTitle }"
          @click="selectPlaylist(playlist.title)"
        >
          <div>
            <strong>{{ playlist.title }}</strong>
            <span>{{ playlist.tracks.length }} 首</span>
          </div>
          <span class="capsule">{{ playlist.tracks.filter((track) => track.audio).length }} ready</span>
        </button>
      </div>
    </aside>

    <main class="player-pane">
      <section class="hero-card">
        <div class="cover-wrap">
          <img
            v-if="activeTrack"
            :src="coverURL(activeTrack.pic)"
            :alt="activeTrack.title"
            class="cover"
          />
          <div v-else class="cover empty">NO TRACK</div>
        </div>

        <div class="hero-copy">
          <p class="eyebrow">Now Playing</p>
          <h2>{{ activeTrack?.title || '选择一首歌开始夜航' }}</h2>
          <p class="subline">{{ activePlaylistTitle || '当前还没有歌单' }}</p>

          <div class="controls">
            <button class="control" @click="playPrevious">Prev</button>
            <button class="control accent" @click="togglePlayback">
              {{ isPlaying ? 'Pause' : 'Play' }}
            </button>
            <button class="control" @click="playNext">Next</button>
            <button class="control" @click="playMode = playMode === 'sequence' ? 'shuffle' : 'sequence'">
              {{ playMode === 'sequence' ? '顺序' : '随机' }}
            </button>
          </div>

          <div class="timeline">
            <span>{{ formatTime(currentTime) }}</span>
            <input
              type="range"
              min="0"
              :max="duration || 0"
              :value="currentTime"
              @input="onSeek"
            />
            <span>{{ formatTime(duration) }}</span>
          </div>

          <div class="volume-strip">
            <span>音量</span>
            <input
              type="range"
              min="0"
              max="1"
              step="0.01"
              :value="volume"
              @input="onVolumeInput"
            />
            <span>{{ Math.round(volume * 100) }}%</span>
          </div>

          <div class="status-line">
            <span>{{ statusText }}</span>
            <span v-if="activeTrack?.audio" class="capsule">audio ready</span>
          </div>
        </div>
      </section>

      <section class="content-grid">
        <div class="panel panel-wide">
          <div class="panel-head">
            <div>
              <p class="eyebrow">Tracks</p>
              <h3>{{ visiblePlaylist?.title || '空歌单' }}</h3>
            </div>
            <button
              v-if="visiblePlaylist"
              class="ghost"
              @click="removePlaylist(visiblePlaylist.title)"
            >
              删除歌单
            </button>
          </div>

          <div v-if="loading" class="panel-body empty-state">正在读取歌单...</div>
          <div v-else-if="error" class="panel-body empty-state error">{{ error }}</div>
          <div v-else-if="!visiblePlaylist" class="panel-body empty-state">没有可展示的歌单。</div>
          <div v-else class="track-list">
            <article
              v-for="track in visiblePlaylist.tracks"
              :key="track.bvid"
              class="track-row"
              :class="{ active: activePlaylistTitle === visiblePlaylist.title && activeTrack?.bvid === track.bvid }"
            >
              <button class="track-main" @click="playTrackByBVID(track.bvid)">
                <img :src="coverURL(track.pic)" :alt="track.title" />
                <span class="track-copy">
                  <strong>{{ track.title }}</strong>
                  <small>{{ track.bvid }}</small>
                </span>
              </button>

              <div class="track-actions">
                <span class="capsule" :class="{ muted: !track.audio }">
                  {{ track.audio ? 'ready' : 'stale' }}
                </span>
                <button class="ghost" @click="downloadCurrentTrack(track)">下载</button>
                <button class="ghost" @click="removeTrack(track)">删除</button>
              </div>
            </article>
          </div>
        </div>
      </section>
    </main>

    <audio
      ref="audioRef"
      preload="auto"
      @loadedmetadata="onLoadedMetadata"
      @timeupdate="onTimeUpdate"
      @ended="onAudioEnded"
      @play="onAudioPlay"
      @pause="onAudioPause"
      @error="onAudioError"
    />
  </div>
</template>
