-- Omnichannel chat: Facebook Messenger + Zalo

CREATE TABLE IF NOT EXISTS chat_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform TEXT NOT NULL CHECK (platform IN ('facebook', 'zalo')),
    name TEXT NOT NULL,
    external_id TEXT,
    cookies_json TEXT,
    zalo_bridge_id TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'error')),
    last_sync_at TIMESTAMPTZ,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_chat_accounts_platform ON chat_accounts(platform);
CREATE INDEX IF NOT EXISTS idx_chat_accounts_status ON chat_accounts(status);

CREATE TABLE IF NOT EXISTS chat_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform TEXT NOT NULL,
    account_id UUID NOT NULL REFERENCES chat_accounts(id) ON DELETE CASCADE,
    thread_id TEXT NOT NULL,
    thread_title TEXT,
    peer_name TEXT,
    peer_avatar TEXT,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    assigned_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    last_message TEXT,
    last_message_at TIMESTAMPTZ,
    unread_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (platform, account_id, thread_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_conversations_account ON chat_conversations(account_id);
CREATE INDEX IF NOT EXISTS idx_chat_conversations_customer ON chat_conversations(customer_id);
CREATE INDEX IF NOT EXISTS idx_chat_conversations_assigned ON chat_conversations(assigned_user_id);
CREATE INDEX IF NOT EXISTS idx_chat_conversations_last_msg ON chat_conversations(last_message_at DESC NULLS LAST);
