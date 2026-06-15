import cors from 'cors'
import express from 'express'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { LoginQRCallbackEventType, ThreadType, Zalo } from 'zca-js'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const PORT = Number(process.env.ZALO_BRIDGE_PORT || 8090)
const DATA_DIR = path.join(__dirname, 'data')
const DATA_FILE = path.join(DATA_DIR, 'accounts.json')

fs.mkdirSync(DATA_DIR, { recursive: true })

/** @type {Record<string, any>} */
let accounts = {}
/** @type {Record<string, import('zca-js').API|null>} */
const apis = {}
/** @type {Record<string, Record<string, any[]>>} */
const messageCache = {}

function loadData() {
  try {
    if (fs.existsSync(DATA_FILE)) {
      accounts = JSON.parse(fs.readFileSync(DATA_FILE, 'utf8'))
    }
  } catch {
    accounts = {}
  }
}

function saveData() {
  fs.writeFileSync(DATA_FILE, JSON.stringify(accounts, null, 2))
}

function genId() {
  return `zalo_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
}

function cacheMessage(accountId, threadId, msg) {
  if (!messageCache[accountId]) messageCache[accountId] = {}
  if (!messageCache[accountId][threadId]) messageCache[accountId][threadId] = []
  const list = messageCache[accountId][threadId]
  const id = msg.message_id || `${msg.timestamp_ms}_${msg.text}`
  if (list.some((m) => m.message_id === id)) return
  list.unshift(msg)
  if (list.length > 200) list.length = 200
}

function attachListener(accountId, api) {
  api.listener.on('message', (message) => {
    const threadId = message.threadId
    const isPlain = typeof message.data?.content === 'string'
    const text = isPlain ? message.data.content : '[Đính kèm]'
    const ownId = api.getOwnId?.() || accounts[accountId]?.own_id || ''
    cacheMessage(accountId, threadId, {
      message_id: message.data?.msgId || `${Date.now()}`,
      sender_fbid: message.data?.uidFrom || '',
      sender_name: message.data?.dName || '',
      text,
      timestamp_ms: String(message.data?.ts || Date.now()),
      is_self: message.data?.uidFrom === ownId,
    })
  })
  api.listener.start?.()
}

async function ensureApi(accountId) {
  if (apis[accountId]) return apis[accountId]
  const acc = accounts[accountId]
  if (!acc?.credentials) return null
  const zalo = new Zalo()
  const api = await zalo.login(acc.credentials)
  apis[accountId] = api
  acc.own_id = api.getOwnId?.() || acc.own_id
  acc.status = 'connected'
  attachListener(accountId, api)
  saveData()
  return api
}

function toInboxThread(threadId, title, avatar, lastMessage, lastMs, isGroup) {
  return {
    thread_id: threadId,
    thread_key: threadId,
    title,
    is_group: isGroup,
    last_activity_ms: String(lastMs || Date.now()),
    last_message: lastMessage || '',
    users: [{ id: threadId, full_name: title, avatar }],
  }
}

const app = express()
app.use(cors())
app.use(express.json({ limit: '5mb' }))

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', accounts: Object.keys(accounts).length })
})

app.post('/api/accounts', (req, res) => {
  const id = genId()
  accounts[id] = {
    id,
    name: req.body?.name || 'Zalo account',
    status: 'pending',
    created_at: new Date().toISOString(),
  }
  saveData()
  res.json({ id })
})

app.post('/api/accounts/:id/login-qr', async (req, res) => {
  const { id } = req.params
  if (!accounts[id]) return res.status(404).json({ error: 'account not found' })

  try {
    const zalo = new Zalo()
    let qrImage = ''
    let qrStatus = 'waiting'

    const loginPromise = zalo.loginQR({ userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/140.0.0.0 Safari/537.36' }, (event) => {
      if (event.type === LoginQRCallbackEventType.QRCodeGenerated) {
        qrImage = event.data.image
        qrStatus = 'qr_ready'
        accounts[id].qr_image = qrImage
        accounts[id].login_status = qrStatus
        saveData()
      }
      if (event.type === LoginQRCallbackEventType.QRCodeScanned) {
        qrStatus = 'scanned'
        accounts[id].login_status = qrStatus
        accounts[id].scanned_name = event.data.display_name
        saveData()
      }
      if (event.type === LoginQRCallbackEventType.GotLoginInfo) {
        accounts[id].credentials = {
          cookie: event.data.cookie,
          imei: event.data.imei,
          userAgent: event.data.userAgent,
        }
        accounts[id].login_status = 'credentials_ready'
        saveData()
      }
    })

    accounts[id].login_status = 'starting'
    saveData()

    loginPromise.then(async (api) => {
      if (!api) {
        accounts[id].status = 'error'
        accounts[id].login_status = 'failed'
        saveData()
        return
      }
      apis[id] = api
      accounts[id].status = 'connected'
      accounts[id].login_status = 'connected'
      accounts[id].own_id = api.getOwnId?.() || ''
      accounts[id].display_name = accounts[id].scanned_name || accounts[id].name
      attachListener(id, api)
      saveData()
    }).catch((err) => {
      accounts[id].status = 'error'
      accounts[id].login_status = 'failed'
      accounts[id].error = String(err)
      saveData()
    })

    // Wait briefly for QR image
    for (let i = 0; i < 20; i++) {
      await new Promise((r) => setTimeout(r, 300))
      if (accounts[id].qr_image) break
    }

    res.json({
      status: accounts[id].login_status || qrStatus,
      qr_image: accounts[id].qr_image || qrImage,
      display_name: accounts[id].scanned_name || null,
    })
  } catch (err) {
    res.status(500).json({ error: String(err) })
  }
})

app.get('/api/accounts/:id/login-status', (req, res) => {
  const acc = accounts[req.params.id]
  if (!acc) return res.status(404).json({ error: 'account not found' })
  res.json({
    status: acc.login_status || acc.status,
    connected: acc.status === 'connected',
    display_name: acc.display_name || acc.scanned_name || acc.name,
    qr_image: acc.qr_image || null,
    error: acc.error || null,
  })
})

app.get('/api/accounts/:id/inbox', async (req, res) => {
  const { id } = req.params
  try {
    const api = await ensureApi(id)
    if (!api) return res.status(400).json({ error: 'account not connected — scan QR first' })

    const threads = []
    const friends = await api.getAllFriends(200, 1)
    for (const f of friends.slice(0, 100)) {
      const tid = f.userId
      const cached = messageCache[id]?.[tid]?.[0]
      threads.push(toInboxThread(
        tid,
        f.displayName || f.zaloName || f.username || tid,
        f.avatar,
        cached?.text || f.status || '',
        f.lastActionTime || cached?.timestamp_ms || Date.now(),
        false,
      ))
    }

    try {
      const groups = await api.getAllGroups()
      for (const g of (groups || []).slice(0, 50)) {
        const tid = g.groupId || g.id
        if (!tid) continue
        const cached = messageCache[id]?.[tid]?.[0]
        threads.push(toInboxThread(
          tid,
          g.name || g.groupName || 'Nhóm Zalo',
          g.avatar || g.fullAvatars?.[0] || '',
          cached?.text || '',
          g.lastMessageTime || cached?.timestamp_ms || Date.now(),
          true,
        ))
      }
    } catch {
      // groups optional
    }

    threads.sort((a, b) => Number(b.last_activity_ms) - Number(a.last_activity_ms))

    res.json({
      threads,
      has_more: false,
      viewer_id: api.getOwnId?.() || accounts[id]?.own_id || '',
    })
  } catch (err) {
    res.status(502).json({ error: String(err) })
  }
})

app.get('/api/accounts/:id/threads/:threadId', async (req, res) => {
  const { id, threadId } = req.params
  const cursor = req.query.cursor || ''
  try {
    const api = await ensureApi(id)
    if (!api) return res.status(400).json({ error: 'account not connected' })

    const ownId = api.getOwnId?.() || accounts[id]?.own_id || ''
    let messages = [...(messageCache[id]?.[threadId] || [])]
    let title = threadId
    let isGroup = false

    try {
      const history = await api.getGroupChatHistory(threadId, 30)
      if (history?.groupMsgs?.length) {
        isGroup = true
        for (const m of history.groupMsgs) {
          messages.push({
            message_id: m.msgId || `${m.ts}`,
            sender_fbid: m.uidFrom || '',
            sender_name: m.dName || '',
            text: m.content || m.msg || '[Đính kèm]',
            timestamp_ms: String(m.ts || Date.now()),
            is_self: m.uidFrom === ownId,
          })
        }
      }
    } catch {
      // not a group — use cache only
    }

    if (!isGroup) {
      try {
        const info = await api.getUserInfo(threadId)
        const u = info?.changed_profiles?.[threadId] || info?.unchanged_profiles?.[threadId]
        if (u) title = u.displayName || u.zaloName || title
      } catch {
        // ignore
      }
    }

    const seen = new Set()
    messages = messages.filter((m) => {
      const key = m.message_id || `${m.timestamp_ms}_${m.text}`
      if (seen.has(key)) return false
      seen.add(key)
      return true
    })
    messages.sort((a, b) => Number(b.timestamp_ms) - Number(a.timestamp_ms))

    res.json({
      thread_id: threadId,
      title,
      users: [{ id: threadId, full_name: title }],
      messages,
      has_more: false,
      next_cursor: cursor,
      viewer_id: ownId,
    })
  } catch (err) {
    res.status(502).json({ error: String(err) })
  }
})

app.post('/api/accounts/:id/send', async (req, res) => {
  const { id } = req.params
  const { thread_id: threadId, text } = req.body || {}
  if (!threadId || !text) return res.status(400).json({ error: 'thread_id and text required' })
  try {
    const api = await ensureApi(id)
    if (!api) return res.status(400).json({ error: 'account not connected' })
    const type = threadId.length > 12 ? ThreadType.Group : ThreadType.User
    await api.sendMessage({ msg: text }, threadId, type)
    const ownId = api.getOwnId?.() || ''
    cacheMessage(id, threadId, {
      message_id: `local_${Date.now()}`,
      sender_fbid: ownId,
      sender_name: 'Tôi',
      text,
      timestamp_ms: String(Date.now()),
      is_self: true,
    })
    res.json({ status: 'sent' })
  } catch (err) {
    res.status(502).json({ error: String(err) })
  }
})

loadData()

// Restore connected sessions on boot
for (const id of Object.keys(accounts)) {
  if (accounts[id].credentials) {
    ensureApi(id).catch(() => {})
  }
}

app.listen(PORT, () => {
  console.log(`zalo-bridge listening on :${PORT}`)
})
