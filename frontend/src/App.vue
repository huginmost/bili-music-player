<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  audioURL,
  coverURL,
  deletePlaylist,
  deleteTrack,
  downloadTrack,
  fetchLibrary,
  prefetchTrackAudio,
  refreshTrackAudio
} from './api'

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

  const result = []
  for (let step = 1; step <= 3; step += 1) {
    const nextIndex = computeNextIndex(step)
    if (nextIndex === -1) {
      break
    }
    result.push(playlist.tracks[nextIndex])
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

function computeNextIndex(offset = 1) {
  const playlist = activePlaylist.value
  if (!playlist || !playlist.tracks.length) {
    return -1
  }

  if (playMode.value === 'shuffle') {
    const choices = playlist.tracks
      .map((_, index) => index)
      .filter((index) => index !== activeTrackIndex.value)
    return choices[Math.floor(Math.random() * choices.length)] ?? -1
  }

  const nextIndex = activeTrackIndex.value + offset
  if (nextIndex >= playlist.tracks.length) {
    return -1
  }

  return nextIndex
}

function syncVolume() {
  if (!audioRef.value) {
    return
  }

  audioRef.value.volume = volume.value
}

async function loadLibrary() {
  loading.value = true
  error.value = ''

  try {
    const payload = await fetchLibrary()
    library.value = payload

    const titles = Object.keys(payload)
    if (!titles.length) {
      activePlaylistTitle.value = ''
      activeTrackIndex.value = 0
      return
    }

    if (!payload[activePlaylistTitle.value]) {
      activePlaylistTitle.value = titles[0]
      activeTrackIndex.value = 0
    } else {
      const trackCount = payload[activePlaylistTitle.value]?.length || 0
      if (activeTrackIndex.value >= trackCount) {
        activeTrackIndex.value = 0
      }
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
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
  await loadLibrary()
  statusText.value = `音频已更新：${refreshed.title}`
  return refreshed
}

async function playTrack(index) {
  const playlist = activePlaylist.value
  if (!playlist || !playlist.tracks[index]) {
    return
  }

  activeTrackIndex.value = index
  error.value = ''

  try {
    const track = await prepareTrack(playlist.tracks[index])
    if (!track || !audioRef.value) {
      return
    }

    audioRef.value.src = audioURL(activePlaylistTitle.value, track.bvid)
    syncVolume()
    await audioRef.value.play()
    isPlaying.value = true
    statusText.value = `正在播放：${track.title}`
  } catch (err) {
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
      await prepareTrack(activeTrack.value)
      audioRef.value.src = audioURL(activePlaylistTitle.value, activeTrack.value.bvid)
      syncVolume()
      await audioRef.value.play()
      isPlaying.value = true
      statusText.value = `正在播放：${activeTrack.value.title}`
    } catch (err) {
      error.value = err.message
    }
    return
  }

  audioRef.value.pause()
  isPlaying.value = false
  statusText.value = '已暂停播放。'
}

async function playNext() {
  const nextIndex = computeNextIndex()
  if (nextIndex === -1) {
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

  const previousIndex = Math.max(activeTrackIndex.value - 1, 0)
  await playTrack(previousIndex)
}

async function queuePrefetch() {
  if (!activePlaylistTitle.value || !upcomingTracks.value.length) {
    return
  }

  const missing = upcomingTracks.value.filter((track) => !track.audio).map((track) => track.bvid)
  if (!missing.length) {
    return
  }

  try {
    await prefetchTrackAudio(activePlaylistTitle.value, missing)
    await loadLibrary()
  } catch (err) {
    error.value = err.message
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
}

function onSeek(event) {
  if (!audioRef.value) {
    return
  }

  audioRef.value.currentTime = Number(event.target.value)
}

function onVolumeInput(event) {
  volume.value = Number(event.target.value)
  syncVolume()
}

function onAudioEnded() {
  playNext()
}

function onKeydown(event) {
  if (event.code === 'Space' && event.target.tagName !== 'INPUT') {
    event.preventDefault()
    togglePlayback()
  }
}

watch(activePlaylistTitle, async () => {
  activeTrackIndex.value = 0
  await queuePrefetch()
})

watch([activeTrackIndex, playMode], async () => {
  await queuePrefetch()
})

onMounted(async () => {
  window.addEventListener('keydown', onKeydown)
  await loadLibrary()
  syncVolume()
  await queuePrefetch()
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
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
      @timeupdate="onTimeUpdate"
      @ended="onAudioEnded"
      @play="isPlaying = true"
      @pause="isPlaying = false"
    />
  </div>
</template>
