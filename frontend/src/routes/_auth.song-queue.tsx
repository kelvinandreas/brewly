import { createRoute } from '@tanstack/react-router'
import { Route as authRoute } from './_auth'
import { useSongRequests } from '../hooks/useSongRequests'
import { useSongQueueSSE } from '../hooks/useSongQueueSSE'
import { getAccessToken } from '../lib/auth'
import type { SongRequest, SongStatus } from '../types/api'

export const Route = createRoute({
  getParentRoute: () => authRoute,
  path: '/song-queue',
  component: SongQueuePage,
})

const COLUMNS: { status: SongStatus; label: string; color: string }[] = [
  { status: 'queued', label: 'Queue', color: 'border-blue-400 bg-blue-50' },
  { status: 'playing', label: 'Now Playing', color: 'border-green-400 bg-green-50' },
  { status: 'played', label: 'Played', color: 'border-gray-300 bg-gray-50' },
  { status: 'skipped', label: 'Skipped', color: 'border-red-200 bg-red-50' },
]

const NEXT_STATUS: Partial<Record<SongStatus, SongStatus>> = {
  queued: 'playing',
  playing: 'played',
}

const NEXT_LABEL: Partial<Record<SongStatus, string>> = {
  queued: 'Play',
  playing: 'Done',
}

function SongQueuePage() {
  const { listQuery, updateStatusMutation } = useSongRequests()
  const initialSongs = listQuery.data ?? []
  const songs = useSongQueueSSE(getAccessToken(), initialSongs)

  const active = songs.filter((s) => s.status === 'queued' || s.status === 'playing')

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <header className="bg-gray-800 px-6 py-4 flex items-center justify-between">
        <h1 className="text-xl font-bold text-purple-400">Song Queue</h1>
        <span className="text-xs text-gray-400">{active.length} active</span>
      </header>

      <div className="p-4 grid grid-cols-2 lg:grid-cols-4 gap-4 items-start">
        {COLUMNS.map((col) => {
          const colSongs = songs.filter((s) => s.status === col.status)
          return (
            <div key={col.status}>
              <div className="flex items-center gap-2 mb-3">
                <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-300">
                  {col.label}
                </h2>
                <span className="bg-gray-700 text-gray-300 text-xs px-1.5 py-0.5 rounded-full">
                  {colSongs.length}
                </span>
              </div>
              <div className="space-y-3">
                {colSongs.map((song) => (
                  <SongCard
                    key={song.id}
                    song={song}
                    colorClass={col.color}
                    nextLabel={NEXT_LABEL[song.status]}
                    onAdvance={() => {
                      const next = NEXT_STATUS[song.status]
                      if (next) updateStatusMutation.mutate({ id: song.id, status: next })
                    }}
                    onSkip={() => updateStatusMutation.mutate({ id: song.id, status: 'skipped' })}
                  />
                ))}
                {colSongs.length === 0 && (
                  <p className="text-xs text-gray-600 text-center py-4">Empty</p>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function SongCard({
  song,
  colorClass,
  nextLabel,
  onAdvance,
  onSkip,
}: {
  song: SongRequest
  colorClass: string
  nextLabel?: string
  onAdvance: () => void
  onSkip: () => void
}) {
  const canAct = song.status === 'queued' || song.status === 'playing'
  return (
    <div className={`rounded-lg border-2 ${colorClass} p-3 text-gray-800`}>
      {song.thumbnailUrl && (
        <img
          src={song.thumbnailUrl}
          alt={song.title}
          className="w-full h-20 object-cover rounded mb-2"
        />
      )}
      <p className="text-sm font-semibold leading-tight line-clamp-2">{song.title}</p>
      <p className="text-xs text-gray-500 mt-0.5">{song.channelName}</p>
      {song.note && <p className="text-xs italic text-gray-400 mt-1">{song.note}</p>}
      {canAct && (
        <div className="flex gap-1 mt-2">
          <button
            onClick={onSkip}
            className="flex-1 text-xs border border-red-300 text-red-500 py-1 rounded hover:bg-red-50"
          >
            Skip
          </button>
          {nextLabel && (
            <button
              onClick={onAdvance}
              className="flex-1 text-xs bg-gray-800 text-white py-1 rounded hover:bg-gray-700"
            >
              {nextLabel}
            </button>
          )}
        </div>
      )}
    </div>
  )
}
