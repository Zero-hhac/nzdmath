import React, { useEffect, useMemo, useRef, useState } from 'react';
import {
  Clock,
  Download,
  FileText,
  Image as ImageIcon,
  LogIn,
  MessageCircle,
  Paperclip,
  RotateCcw,
  Send,
  Users,
  X,
} from 'lucide-react';
import type { ViewProps } from '@/src/types/app';
import { api } from '@/src/lib/api';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';
import { LoginModal } from '@/src/components/LoginModal';

type ChatMessage = {
  id: number;
  user_id: number;
  user_name?: string;
  user_avatar?: string;
  message_type: 'text' | 'image' | 'file' | 'system';
  content?: string;
  file_name?: string;
  file_url?: string;
  file_size?: number;
  file_ext?: string;
  created_at: string;
};

const maxFileSize = 5 * 1024 * 1024;
const allowedExts = [
  '.jpg', '.jpeg', '.png', '.gif', '.webp',
  '.txt', '.md', '.py', '.js', '.ts', '.json', '.csv',
  '.pdf', '.doc', '.docx', '.ppt', '.pptx',
];
const videoExts = ['.mp4', '.mov', '.avi', '.mkv', '.webm', '.flv', '.wmv'];
const recallWindowMs = 2 * 60 * 1000;

function formatTime(value?: string) {
  if (!value) return '';
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatSize(bytes?: number) {
  if (!bytes) return '';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

function fileExt(name: string) {
  const dot = name.lastIndexOf('.');
  return dot >= 0 ? name.slice(dot).toLowerCase() : '';
}

function assetUrl(url?: string) {
  if (!url) return '';
  return url;
}

export const ChatView: React.FC<ViewProps> = ({ navigate }) => {
  const { user, loading: authLoading } = useAuth();
  const { showToast } = useToast();
  const [loginOpen, setLoginOpen] = useState(false);
  const [joined, setJoined] = useState(false);
  const [joining, setJoining] = useState(false);
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [loadingEarlier, setLoadingEarlier] = useState(false);
  const [hasEarlier, setHasEarlier] = useState(false);
  const [beforeId, setBeforeId] = useState<number>();
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [onlineCount, setOnlineCount] = useState(0);
  const [content, setContent] = useState('');
  const [sending, setSending] = useState(false);
  const [uploading, setUploading] = useState(false);
  const fileRef = useRef<HTMLInputElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const joinedRef = useRef(false);
  const lastMessageIdRef = useRef(0);
  const lastDeleteMsRef = useRef(0);

  const lastMessageId = useMemo(() => {
    return messages.reduce((max, msg) => Math.max(max, msg.id), 0);
  }, [messages]);

  useEffect(() => {
    lastMessageIdRef.current = lastMessageId;
  }, [lastMessageId]);

  useEffect(() => {
    joinedRef.current = joined;
  }, [joined]);

  useEffect(() => {
    return () => {
      if (joinedRef.current) {
        api.chatLeave().catch(() => {});
      }
    };
  }, []);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages.length]);

  const mergeMessages = (incoming: ChatMessage[]) => {
    if (!incoming.length) return;
    setMessages((current) => {
      const map = new Map<number, ChatMessage>();
      current.forEach((msg) => map.set(msg.id, msg));
      incoming.forEach((msg) => map.set(msg.id, msg));
      return Array.from(map.values()).sort((a, b) => {
        const timeDelta = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
        return timeDelta || a.id - b.id;
      });
    });
  };

  const applyDeletedMessages = (ids?: number[]) => {
    if (!ids?.length) return;
    const deleted = new Set(ids);
    setMessages((current) => current.filter((message) => !deleted.has(message.id)));
  };

  const loadHistory = async () => {
    setLoadingMessages(true);
    try {
      const res = await api.getChatMessages({ limit: 50 });
      const items = (res.data.messages || []) as ChatMessage[];
      setMessages(items.sort((a, b) => {
        const timeDelta = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
        return timeDelta || a.id - b.id;
      }));
      setOnlineCount(res.data.online_count || 0);
      setHasEarlier(Boolean(res.data.has_more));
      setBeforeId(res.data.next_before_id);
      lastDeleteMsRef.current = res.data.deleted_at_ms || Date.now();
    } finally {
      setLoadingMessages(false);
    }
  };

  const loadEarlier = async () => {
    if (!beforeId || !hasEarlier || loadingEarlier) return;
    setLoadingEarlier(true);
    try {
      const res = await api.getChatMessages({ beforeId, limit: 50 });
      mergeMessages((res.data.messages || []) as ChatMessage[]);
      setHasEarlier(Boolean(res.data.has_more));
      setBeforeId(res.data.next_before_id);
      applyDeletedMessages(res.data.deleted_ids);
      if (res.data.deleted_at_ms) lastDeleteMsRef.current = res.data.deleted_at_ms;
    } catch (err: any) {
      showToast(err.message || '更早消息加载失败', 'error');
    } finally {
      setLoadingEarlier(false);
    }
  };

  const joinChat = async () => {
    setJoining(true);
    try {
      const joinedRes = await api.chatJoin();
      setOnlineCount(joinedRes.data.online_count || 0);
      setJoined(true);
      await loadHistory();
    } catch (err: any) {
      showToast(err.message || '加入聊天室失败', 'error');
    } finally {
      setJoining(false);
    }
  };

  const leaveChat = async () => {
    try {
      await api.chatLeave();
    } catch {}
    setJoined(false);
    setOnlineCount(0);
  };

  useEffect(() => {
    if (!joined) return undefined;
    let timer: number | undefined;
    let stopped = false;
    const controller = new AbortController();
    const poll = async () => {
      try {
        const res = await api.getChatMessages({
          afterId: lastMessageIdRef.current || undefined,
          limit: 100,
          afterDeleteMs: lastDeleteMsRef.current || undefined,
        }, controller.signal);
        if (stopped) return;
        mergeMessages((res.data.messages || []) as ChatMessage[]);
        applyDeletedMessages(res.data.deleted_ids);
        setOnlineCount(res.data.online_count || 0);
        if (res.data.deleted_at_ms) lastDeleteMsRef.current = res.data.deleted_at_ms;
      } catch (err: any) {
        if (stopped || err?.name === 'AbortError') return;
        setJoined(false);
        showToast(err.message || '聊天室连接已断开', 'error');
        return;
      }
      if (!stopped) {
        timer = window.setTimeout(poll, document.hidden ? 10000 : 2000);
      }
    };
    timer = window.setTimeout(poll, 1000);

    return () => {
      stopped = true;
      controller.abort();
      if (timer) window.clearTimeout(timer);
    };
  }, [joined, showToast]);

  const recallMessage = async (message: ChatMessage) => {
    try {
      await api.deleteChatMessage(message.id);
      setMessages((current) => current.filter((item) => item.id !== message.id));
      showToast('消息已撤回', 'success');
    } catch (err: any) {
      showToast(err.message || '撤回失败', 'error');
    }
  };

  const sendText = async (e?: React.FormEvent) => {
    e?.preventDefault();
    const text = content.trim();
    if (!text || sending) return;
    setSending(true);
    try {
      const res = await api.sendChatText(text);
      mergeMessages([res.data as ChatMessage]);
      setContent('');
    } catch (err: any) {
      showToast(err.message || '发送失败', 'error');
    } finally {
      setSending(false);
    }
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;

    const ext = fileExt(file.name);
    if (file.size > maxFileSize) {
      showToast('文件不能超过 5MB', 'error');
      return;
    }
    if (videoExts.includes(ext) || file.type.startsWith('video/')) {
      showToast('不支持上传视频文件', 'error');
      return;
    }
    if (!allowedExts.includes(ext)) {
      showToast('不支持的文件格式', 'error');
      return;
    }

    setUploading(true);
    try {
      const fd = new FormData();
      fd.append('file', file);
      const res = await api.sendChatFile(fd);
      mergeMessages([res.data as ChatMessage]);
    } catch (err: any) {
      showToast(err.message || '上传失败', 'error');
    } finally {
      setUploading(false);
    }
  };

  if (authLoading) {
    return <div className="text-center text-zinc-500 py-20">加载中...</div>;
  }

  if (!user) {
    return (
      <>
        <div className="space-y-8">
          <div className="page-intro space-y-2">
            <div className="section-kicker">Chat Room</div>
            <h2 className="section-title">聊天室</h2>
            <p className="section-subtitle">会员登录后进入</p>
          </div>
          <div className="sidebar-panel rounded-[2rem] p-10 text-center space-y-6">
            <div className="w-16 h-16 rounded-2xl bg-primary/10 mx-auto flex items-center justify-center">
              <MessageCircle className="w-8 h-8 text-primary" />
            </div>
            <button onClick={() => setLoginOpen(true)} className="btn-primary mx-auto">
              <LogIn className="w-4 h-4" />
              登录
            </button>
          </div>
        </div>
        <LoginModal open={loginOpen} onClose={() => setLoginOpen(false)} />
      </>
    );
  }

  if (!joined) {
    return (
      <div className="space-y-8">
        <div className="page-intro flex flex-col md:flex-row md:items-end md:justify-between gap-5">
          <div className="space-y-2">
            <div className="section-kicker">Chat Room</div>
            <h2 className="section-title">聊天室</h2>
            <p className="section-subtitle">欢迎，{user.nickname || user.username}</p>
          </div>
          <button onClick={() => navigate('portal')} className="btn-secondary !py-2.5 !text-xs">
            会员中心
          </button>
        </div>

        <div className="sidebar-panel rounded-[2rem] p-10 text-center space-y-6">
          <div className="w-16 h-16 rounded-2xl bg-pastel-blue mx-auto flex items-center justify-center">
            <Users className="w-8 h-8 text-pastel-blue-text" />
          </div>
          <button onClick={joinChat} disabled={joining} className="btn-primary mx-auto">
            <MessageCircle className="w-4 h-4" />
            {joining ? '加入中...' : '加入聊天室'}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="page-intro flex flex-col md:flex-row md:items-end md:justify-between gap-5">
        <div className="space-y-2">
          <div className="section-kicker">Chat Room</div>
          <h2 className="section-title">聊天室</h2>
          <p className="section-subtitle flex items-center gap-2">
            <Users className="w-4 h-4" />
            在线 {onlineCount} 人
          </p>
        </div>
        <button onClick={leaveChat} className="btn-secondary !py-2.5 !text-xs">
          <X className="w-4 h-4" />
          离开聊天室
        </button>
      </div>

      <div className="glass-card rounded-2xl overflow-hidden">
        <div className="h-[52vh] min-h-[320px] md:h-[58vh] md:min-h-[460px] overflow-y-auto bg-canvas-alt/70 px-3 py-5 sm:px-4 md:px-6">
          {loadingMessages ? (
            <div className="text-center text-zinc-500 py-20">消息加载中...</div>
          ) : messages.length === 0 ? (
            <div className="text-center text-zinc-500 py-20">暂无消息</div>
          ) : (
            <div className="space-y-4">
              {hasEarlier && (
                <div className="flex justify-center pb-2">
                  <button onClick={loadEarlier} disabled={loadingEarlier} className="btn-ghost !px-4 !py-2 !text-xs">
                    <RotateCcw className={`h-3.5 w-3.5 ${loadingEarlier ? 'animate-spin' : ''}`} />
                    {loadingEarlier ? '加载中' : '查看更早消息'}
                  </button>
                </div>
              )}
              {messages.map((msg) => (
                <MessageItem
                  key={msg.id}
                  message={msg}
                  mine={msg.user_id === user.id}
                  onRecall={() => recallMessage(msg)}
                />
              ))}
              <div ref={bottomRef} />
            </div>
          )}
        </div>

        <form onSubmit={sendText} className="border-t border-border bg-surface p-4">
          <input
            ref={fileRef}
            type="file"
            accept={allowedExts.join(',')}
            onChange={handleFileChange}
            className="hidden"
          />
          <div className="flex items-end gap-2 sm:gap-3">
            <button
              type="button"
              onClick={() => fileRef.current?.click()}
              disabled={uploading}
              className="btn-secondary !px-3 !py-3 shrink-0"
              title="上传图片或文件"
              aria-label="上传图片或文件"
            >
              {uploading ? <ImageIcon className="w-4 h-4 animate-pulse" /> : <Paperclip className="w-4 h-4" />}
            </button>
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault();
                  sendText();
                }
              }}
              rows={2}
              maxLength={2000}
              className="app-input min-h-[48px] min-w-0 max-h-32 flex-1 resize-y rounded-2xl px-3 sm:px-4 py-3"
              placeholder="输入消息..."
            />
            <button type="submit" disabled={!content.trim() || sending} className="btn-primary !px-4 !py-3 shrink-0 !bg-[#4c9eeb] hover:!bg-[#3a82d8]" title="发送" aria-label="发送消息">
              <Send className="w-4 h-4" />
              <span className="hidden sm:inline">{sending ? '发送中' : '发送'}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

function MessageItem({ message, mine, onRecall }: { message: ChatMessage; mine: boolean; onRecall: () => void }) {
  if (message.message_type === 'system') {
    return (
      <div className="flex justify-center">
        <div className="rounded-full bg-black/[0.04] px-3 py-1 text-xs text-text-muted">
          {message.content}
        </div>
      </div>
    );
  }

  const name = message.user_name || `用户 ${message.user_id}`;
  const avatarText = name.slice(0, 1).toUpperCase();
  const canRecall = mine && Date.now() - new Date(message.created_at).getTime() <= recallWindowMs;

  return (
    <div className={`flex gap-3 ${mine ? 'justify-end' : 'justify-start'}`}>
      {!mine && (
        message.user_avatar ? (
          <img src={assetUrl(message.user_avatar)} alt="" className="mt-5 h-9 w-9 rounded-full object-cover border border-border" />
        ) : (
          <div className="mt-5 h-9 w-9 rounded-full bg-pastel-blue flex items-center justify-center text-pastel-blue-text text-xs font-bold">
            {avatarText}
          </div>
        )
      )}

      <div className={`max-w-[78%] space-y-1 ${mine ? 'items-end text-right' : ''}`}>
        <div className={`flex items-center gap-2 text-[11px] text-text-muted ${mine ? 'justify-end' : ''}`}>
          {canRecall && (
            <button onClick={onRecall} className="inline-flex items-center gap-1 hover:text-accent" title="两分钟内可撤回">
              <RotateCcw className="h-3 w-3" />
              撤回
            </button>
          )}
          <span className="font-medium text-charcoal">{name}</span>
          <span className="inline-flex items-center gap-1">
            <Clock className="w-3 h-3" />
            {formatTime(message.created_at)}
          </span>
        </div>
        <div className={`rounded-2xl px-4 py-3 text-sm leading-relaxed shadow-sm ${
          mine
            ? 'bg-[#4c9eeb] text-white rounded-tr-md selection:bg-white/30 selection:text-white'
            : 'bg-white border border-border text-charcoal rounded-tl-md selection:bg-accent/10 selection:text-charcoal'
        }`}>
          {message.message_type === 'text' && <div className="whitespace-pre-wrap break-words">{message.content}</div>}
          {message.message_type === 'image' && (
            <a href={assetUrl(message.file_url)} target="_blank" rel="noreferrer" className="block">
              <img
                src={assetUrl(message.file_url)}
                alt={message.file_name || '聊天图片'}
                className="max-h-72 max-w-full rounded-xl object-contain"
              />
              <div className={`mt-2 text-xs ${mine ? 'text-white/70' : 'text-text-muted'}`}>{message.file_name}</div>
            </a>
          )}
          {message.message_type === 'file' && (
            <a
              href={assetUrl(message.file_url)}
              target="_blank"
              rel="noreferrer"
              className={`flex min-w-0 items-center gap-3 rounded-xl p-3 ${
                mine ? 'bg-white/10 hover:bg-white/15' : 'bg-canvas-alt hover:bg-black/[0.03]'
              }`}
            >
              <FileText className="h-8 w-8 shrink-0" />
              <span className="min-w-0 flex-1 text-left">
                <span className="block truncate font-medium">{message.file_name || '文件'}</span>
                <span className={`block text-xs ${mine ? 'text-white/70' : 'text-text-muted'}`}>{formatSize(message.file_size)}</span>
              </span>
              <Download className="h-4 w-4 shrink-0" />
            </a>
          )}
        </div>
      </div>
    </div>
  );
}
