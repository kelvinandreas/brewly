import { useEffect, useReducer } from 'react'
import type { SongRequest, SongQueueSSEEvent } from '../types/api'

type State = { songs: SongRequest[] }
type Action =
  | { type: 'init'; songs: SongRequest[] }
  | { type: 'song.requested'; song: SongRequest }
  | { type: 'song.status_changed'; song: SongRequest }

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'init':
      return { songs: action.songs }
    case 'song.requested':
      return { songs: [action.song, ...state.songs] }
    case 'song.status_changed':
      return {
        songs: state.songs.map((s) => (s.id === action.song.id ? action.song : s)),
      }
  }
}

export function useSongQueueSSE(accessToken: string | null, initialSongs: SongRequest[] = []) {
  const [state, dispatch] = useReducer(reducer, { songs: initialSongs })

  useEffect(() => {
    dispatch({ type: 'init', songs: initialSongs })
  }, [initialSongs]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!accessToken) return

    const es = new EventSource(`/api/sse/song-queue?token=${encodeURIComponent(accessToken)}`)

    const handleEvent = (e: MessageEvent) => {
      try {
        const evt = JSON.parse(e.data) as SongQueueSSEEvent
        dispatch({ type: evt.type, song: evt.payload })
      } catch {
        // ignore malformed events
      }
    }

    es.addEventListener('song.requested', handleEvent as EventListener)
    es.addEventListener('song.status_changed', handleEvent as EventListener)

    return () => es.close()
  }, [accessToken])

  return state.songs
}
