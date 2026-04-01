import { ref, onUnmounted } from 'vue'

export interface SSEEvent {
  type: string
  data: unknown
}

export function useSSE(url: string) {
  const events = ref<SSEEvent[]>([])
  const connected = ref(false)
  const error = ref<string | null>(null)
  let eventSource: EventSource | null = null

  const eventTypes = [
    'run_started',
    'trial_started',
    'trial_completed',
    'run_progress',
    'log',
    'run_completed',
    'run_error',
  ]

  function connect() {
    if (eventSource) {
      eventSource.close()
    }

    eventSource = new EventSource(url)

    eventSource.onopen = () => {
      connected.value = true
      error.value = null
    }

    eventSource.onerror = () => {
      connected.value = false
      error.value = 'SSE connection lost'
    }

    for (const type of eventTypes) {
      eventSource.addEventListener(type, (e: MessageEvent) => {
        try {
          const data = JSON.parse(e.data)
          events.value.push({ type, data })
        } catch {
          events.value.push({ type, data: e.data })
        }
      })
    }
  }

  function close() {
    if (eventSource) {
      eventSource.close()
      eventSource = null
      connected.value = false
    }
  }

  onUnmounted(() => {
    close()
  })

  return { events, connected, error, connect, close }
}
