// ─── State ─────────────────────────────────────────────
const state = {
  items: [],
  offset: 0,
  hasMore: true,
  loading: false,
  searchQuery: '',
}
let searchTimer

// ─── API ───────────────────────────────────────────────
async function apiGet(path) {
  const res = await fetch(path)
  const data = await res.json()
  if (!res.ok) throw { status: res.status, ...data }
  return data
}

async function apiPost(path, body) {
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  const data = await res.json()
  if (!res.ok) throw { status: res.status, ...data }
  return data
}

// ─── Router ────────────────────────────────────────────
function navigate(path) {
  history.pushState(null, '', path)
  render()
}

window.addEventListener('popstate', render)
document.addEventListener('click', (e) => {
  const link = e.target.closest('[data-nav]')
  if (!link) return
  e.preventDefault()
  navigate(link.getAttribute('href'))
})

// ─── Render ────────────────────────────────────────────
function render() {
  const path = location.pathname
  const m = path.match(/^\/items\/(\d+)/)
  if (m) {
    renderItemDetail(Number(m[1]))
  } else {
    renderItemList()
  }
}

// ─── Item List Page ────────────────────────────────────
function renderItemList() {
  const main = document.getElementById('main')
  main.innerHTML = `
    <div class="card">
      <form id="add-form">
        <div class="form-row">
          <div class="form-group">
            <label for="url-input">Steam Market URL</label>
            <input id="url-input" type="text" class="form-input" placeholder="https://steamcommunity.com/market/listings/730/..." />
          </div>
          <button type="submit" class="btn btn-primary">Add Item</button>
        </div>
      </form>
      <div id="add-msg"></div>
    </div>
    <input id="search-input" type="text" class="form-input search-input"
           placeholder="Search items..." value="${esc(state.searchQuery)}" />
    <div id="items-list"></div>
    <div id="list-bottom"></div>
  `

  document.getElementById('add-form').addEventListener('submit', handleAddItem)
  document.getElementById('search-input').addEventListener('input', onSearchInput)

  state.offset = 0
  state.hasMore = true
  fetchItems()
}

async function fetchItems() {
  state.loading = true
  const list = document.getElementById('items-list')
  if (state.offset === 0) list.innerHTML = '<div class="spinner"></div>'

  try {
    const params = new URLSearchParams({ offset: state.offset, limit: 20 })
    if (state.searchQuery) params.set('q', state.searchQuery)
    const data = await apiGet(`/api/items?${params}`)
    if (state.offset === 0) {
      state.items = data.items
    } else {
      state.items = state.items.concat(data.items)
    }
    state.hasMore = data.items.length === 20

    list.innerHTML = state.items.length === 0
      ? '<div class="empty-state"><h2>No items tracked yet</h2><p>Add a Steam Market URL above to start tracking prices.</p></div>'
      : state.items.map(renderItemCard).join('')

    renderLoadMore()
  } catch {
    list.innerHTML = '<div class="msg msg-error">Failed to load items. Is the server running?</div>'
  } finally {
    state.loading = false
  }
}

function renderItemCard(item) {
  return `
    <div class="card">
      <div class="card-header">
        <div>
          <a href="/items/${item.id}" class="card-title" data-nav>${esc(item.name)}</a>
          <div class="card-url">${esc(item.url)}</div>
        </div>
        <a href="/items/${item.id}" class="btn" data-nav>View History</a>
      </div>
    </div>
  `
}

function renderLoadMore() {
  const bottom = document.getElementById('list-bottom')
  if (!state.hasMore) { bottom.innerHTML = ''; return }
  bottom.innerHTML = `
    <div class="pagination">
      <button id="load-more" class="btn" ${state.loading ? 'disabled' : ''}>
        ${state.loading ? 'Loading...' : 'Load More'}
      </button>
    </div>
  `
  document.getElementById('load-more')?.addEventListener('click', () => {
    state.offset += 20
    fetchItems()
  })
}

async function handleAddItem(e) {
  e.preventDefault()
  const input = document.getElementById('url-input')
  const msg = document.getElementById('add-msg')
  const url = input.value.trim()
  if (!url) return

  const btn = e.target.querySelector('button[type="submit"]')
  btn.disabled = true
  btn.textContent = 'Adding...'
  msg.innerHTML = ''

  try {
    await apiPost('/api/items', { url })
    msg.innerHTML = '<div class="msg msg-success">Item added successfully!</div>'
    input.value = ''
    state.offset = 0
    await fetchItems()
  } catch (err) {
    if (err.status === 409) {
      msg.innerHTML = '<div class="msg msg-success">Item already in your list.</div>'
      input.value = ''
      state.offset = 0
      await fetchItems()
    } else {
      msg.innerHTML = `<div class="msg msg-error">${esc(err.error || 'Failed to add item.')}</div>`
    }
  } finally {
    btn.disabled = false
    btn.textContent = 'Add Item'
  }
}

// ─── Search ────────────────────────────────────────────
function onSearchInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    state.searchQuery = document.getElementById('search-input').value.trim()
    state.offset = 0
    fetchItems()
  }, 300)
}

// ─── Item Detail Page ──────────────────────────────────
async function renderItemDetail(id) {
  const main = document.getElementById('main')
  main.innerHTML = '<div class="spinner"></div>'

  try {
    const [item, history] = await Promise.all([
      apiGet(`/api/items/${id}`),
      apiGet(`/api/items/${id}/history?interval=hour&mode=last&offset=0&limit=100`),
    ])

    main.innerHTML = `
      <a href="/" class="back-link" data-nav>&larr; Back to items</a>
      <div class="detail-name">${esc(item.name)}</div>
      <div class="detail-url">${esc(item.url)}</div>
      <div class="card">
        <div class="controls">
          <div class="form-group">
            <label for="chart-interval">Interval</label>
            <select id="chart-interval">
              <option value="hour">Hourly</option>
              <option value="day">Daily</option>
              <option value="week">Weekly</option>
            </select>
          </div>
          <div class="form-group">
            <label for="chart-mode">Aggregation</label>
            <select id="chart-mode">
              <option value="last">Last Price</option>
              <option value="avg">Average Price</option>
            </select>
          </div>
        </div>
        <div class="chart-container" id="chart-container">
          <div class="chart-empty">No price data available yet.</div>
        </div>
      </div>
    `

    document.getElementById('chart-interval').addEventListener('change', () => reloadChart(id))
    document.getElementById('chart-mode').addEventListener('change', () => reloadChart(id))

    if (history.prices && history.prices.length > 0) {
      drawChart(history.prices, history.interval, history.mode)
    }
  } catch {
    main.innerHTML = '<div class="msg msg-error">Failed to load item details.</div>'
  }
}

async function reloadChart(itemId) {
  const interval = document.getElementById('chart-interval').value
  const mode = document.getElementById('chart-mode').value
  const container = document.getElementById('chart-container')
  container.innerHTML = '<div class="spinner"></div>'

  try {
    const history = await apiGet(`/api/items/${itemId}/history?interval=${interval}&mode=${mode}&offset=0&limit=100`)
    if (history.prices && history.prices.length > 0) {
      drawChart(history.prices, interval, mode)
    } else {
      container.innerHTML = '<div class="chart-empty">No price data available yet.</div>'
    }
  } catch {
    container.innerHTML = '<div class="chart-empty">Failed to load chart data.</div>'
  }
}

// ─── Canvas Chart ──────────────────────────────────────
function drawChart(prices, interval) {
  const container = document.getElementById('chart-container')
  const rect = container.getBoundingClientRect()
  const w = Math.max(rect.width - 0, 300)
  const h = 400

  container.innerHTML = ''
  const canvas = document.createElement('canvas')
  canvas.width = w * 2
  canvas.height = h * 2
  canvas.style.width = w + 'px'
  canvas.style.height = h + 'px'
  container.appendChild(canvas)

  const ctx = canvas.getContext('2d')
  ctx.scale(2, 2)

  const pad = { top: 20, right: 20, bottom: 40, left: 60 }
  const pw = w - pad.left - pad.right
  const ph = h - pad.top - pad.bottom

  const values = prices.map(p => p.price)
  const times = prices.map(p => new Date(p.time))
  const minVal = Math.min(...values)
  const maxVal = Math.max(...values)
  const valRange = maxVal - minVal || 1
  const padding = valRange * 0.1
  const yMin = minVal - padding
  const yMax = maxVal + padding
  const yRange = yMax - yMin

  // Grid lines
  ctx.strokeStyle = '#30363d'
  ctx.lineWidth = 1
  const gridCount = 5
  for (let i = 0; i <= gridCount; i++) {
    const y = pad.top + (ph / gridCount) * i
    ctx.beginPath()
    ctx.moveTo(pad.left, y)
    ctx.lineTo(w - pad.right, y)
    ctx.stroke()
  }

  // Y axis labels
  ctx.fillStyle = '#8b949e'
  ctx.font = '11px sans-serif'
  ctx.textAlign = 'right'
  ctx.textBaseline = 'middle'
  for (let i = 0; i <= gridCount; i++) {
    const y = pad.top + (ph / gridCount) * i
    const val = yMax - (yRange / gridCount) * i
    ctx.fillText('$' + val.toFixed(2), pad.left - 8, y)
  }

  // X axis labels
  ctx.textAlign = 'center'
  ctx.textBaseline = 'top'
  const labelCount = Math.min(times.length, 6)
  const step = Math.max(1, Math.floor((times.length - 1) / (labelCount - 1)))
  for (let i = 0; i < labelCount; i++) {
    const idx = Math.min(i * step, times.length - 1)
    const x = pad.left + (pw / (times.length - 1 || 1)) * idx
    const label = formatTime(times[idx], interval)
    ctx.fillText(label, x, h - pad.bottom + 8)
  }

  // Line
  if (prices.length < 2) {
    // Single point
    const x = pad.left + pw / 2
    const y = pad.top + ph - ((values[0] - yMin) / yRange) * ph
    ctx.fillStyle = '#58a6ff'
    ctx.beginPath()
    ctx.arc(x, y, 4, 0, Math.PI * 2)
    ctx.fill()
    return
  }

  ctx.strokeStyle = '#58a6ff'
  ctx.lineWidth = 2
  ctx.lineJoin = 'round'
  ctx.lineCap = 'round'
  ctx.beginPath()
  for (let i = 0; i < prices.length; i++) {
    const x = pad.left + (pw / (prices.length - 1)) * i
    const y = pad.top + ph - ((values[i] - yMin) / yRange) * ph
    i === 0 ? ctx.moveTo(x, y) : ctx.lineTo(x, y)
  }
  ctx.stroke()

  // Gradient fill under line
  const gradient = ctx.createLinearGradient(0, pad.top, 0, pad.top + ph)
  gradient.addColorStop(0, 'rgba(88, 166, 255, 0.15)')
  gradient.addColorStop(1, 'rgba(88, 166, 255, 0.01)')
  ctx.fillStyle = gradient
  ctx.beginPath()
  ctx.moveTo(pad.left, pad.top + ph)
  for (let i = 0; i < prices.length; i++) {
    const x = pad.left + (pw / (prices.length - 1)) * i
    const y = pad.top + ph - ((values[i] - yMin) / yRange) * ph
    ctx.lineTo(x, y)
  }
  ctx.lineTo(pad.left + pw, pad.top + ph)
  ctx.closePath()
  ctx.fill()

  // Tooltip on hover
  let tooltipEl = container.querySelector('.chart-tooltip')
  if (!tooltipEl) {
    tooltipEl = document.createElement('div')
    tooltipEl.className = 'chart-tooltip'
    tooltipEl.style.cssText = `
      position: absolute; display: none; background: #161b22; border: 1px solid #30363d;
      border-radius: 6px; padding: 8px 12px; font-size: 13px; color: #e1e4e8;
      pointer-events: none; white-space: nowrap;
    `
    container.appendChild(tooltipEl)
    container.style.position = 'relative'
  }

  canvas.addEventListener('mousemove', (e) => {
    const r = canvas.getBoundingClientRect()
    const mx = e.clientX - r.left
    const idx = Math.round(((mx - pad.left) / pw) * (prices.length - 1))
    if (idx < 0 || idx >= prices.length) { tooltipEl.style.display = 'none'; return }

    const x = pad.left + (pw / (prices.length - 1)) * idx
    const y = pad.top + ph - ((values[idx] - yMin) / yRange) * ph

    tooltipEl.style.display = 'block'
    tooltipEl.style.left = (x + 12) + 'px'
    tooltipEl.style.top = (y - 30) + 'px'
    tooltipEl.innerHTML = `
      <div>$${values[idx].toFixed(2)}</div>
      <div style="color:#8b949e;font-size:11px">${formatTime(times[idx], interval)}</div>
    `
  })

  canvas.addEventListener('mouseleave', () => {
    tooltipEl.style.display = 'none'
  })
}

function formatTime(date, interval) {
  if (interval === 'hour') {
    return date.toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit' })
  }
  if (interval === 'day') {
    return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
  }
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: '2-digit' })
}

// ─── Helpers ──────────────────────────────────────────
function esc(str) {
  const div = document.createElement('div')
  div.textContent = str
  return div.innerHTML
}

// ─── Init ──────────────────────────────────────────────
render()
