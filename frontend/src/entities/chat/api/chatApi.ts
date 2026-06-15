import { http } from '@/shared/api/http'

export interface ChatAccount {
  id: string
  platform: 'facebook' | 'zalo'
  name: string
  external_id?: string
  status: string
  has_credentials: boolean
  last_sync_at?: string
  created_at: string
}

export interface ChatUser {
  id: string
  full_name: string
  avatar?: string
}

export interface ChatThread {
  thread_id: string
  thread_key: string
  title: string
  is_group: boolean
  last_message: string
  last_activity_ms: string
  platform?: string
  account_id?: string
  conversation_id?: string
  customer_id?: string
  customer_name?: string
  assigned_user_id?: string
  assigned_name?: string
  users?: ChatUser[]
}

export interface ChatMessage {
  message_id: string
  sender_fbid: string
  sender_name: string
  sender_avatar?: string
  text: string
  timestamp_ms: string
  is_self?: boolean
  content_type?: string
}

export interface InboxResponse {
  threads: ChatThread[]
  has_more: boolean
  viewer_id?: string
}

export interface ThreadResponse {
  thread_id: string
  title: string
  messages: ChatMessage[]
  has_more: boolean
  next_cursor?: string
  viewer_id?: string
}

export interface ConversationMeta {
  id: string
  platform: string
  account_id: string
  thread_id: string
  customer_id?: string
  customer_name?: string
  assigned_user_id?: string
  assigned_name?: string
}

export const chatApi = {
  listAccounts: (platform?: string): Promise<ChatAccount[]> => {
    const q = platform ? `?platform=${platform}` : ''
    return http(`/api/v1/chat/accounts${q}`)
  },
  createAccount: (body: { platform: string; name: string; cookies_json?: string }): Promise<ChatAccount> =>
    http('/api/v1/chat/accounts', { method: 'POST', body }),
  updateAccount: (id: string, body: { name?: string; cookies_json?: string }): Promise<ChatAccount> =>
    http(`/api/v1/chat/accounts/${id}`, { method: 'PUT', body }),
  deleteAccount: (id: string): Promise<{ message: string }> =>
    http(`/api/v1/chat/accounts/${id}`, { method: 'DELETE' }),
  zaloQR: (id: string): Promise<{ status: string; qr_image?: string; display_name?: string }> =>
    http(`/api/v1/chat/accounts/${id}/zalo/qr`, { method: 'POST' }),
  zaloStatus: (id: string): Promise<{ status: string; connected: boolean; qr_image?: string; display_name?: string }> =>
    http(`/api/v1/chat/accounts/${id}/zalo/status`),
  inbox: (accountId: string): Promise<InboxResponse> =>
    http(`/api/v1/chat/inbox?account_id=${encodeURIComponent(accountId)}`),
  thread: (accountId: string, threadId: string, cursor?: string): Promise<ThreadResponse> => {
    const q = new URLSearchParams({ account_id: accountId })
    if (cursor) q.set('cursor', cursor)
    return http(`/api/v1/chat/threads/${encodeURIComponent(threadId)}?${q}`)
  },
  send: (body: { account_id: string; thread_id: string; text: string }): Promise<{ status: string }> =>
    http('/api/v1/chat/send', { method: 'POST', body }),
  getConversation: (id: string): Promise<ConversationMeta> =>
    http(`/api/v1/chat/conversations/${id}`),
  updateConversation: (id: string, body: { customer_id?: string; assigned_user_id?: string }): Promise<ConversationMeta> =>
    http(`/api/v1/chat/conversations/${id}`, { method: 'PUT', body }),
  assignees: (): Promise<{ id: string; full_name: string }[]> =>
    http('/api/v1/chat/assignees'),
}
