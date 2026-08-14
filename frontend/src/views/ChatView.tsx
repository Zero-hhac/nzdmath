import React, { useEffect, useMemo, useRef, useState, useCallback } from 'react';
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
  Headphones,
  ShieldCheck,
} from 'lucide-react';
import type { ViewProps } from '@/src/types/app';
import { api } from '@/src/lib/api';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';
import { connectChatWS } from '@/src/lib/ws';
import { authFetchBlob } from '@/src/lib/http';
import { LoginModal } from '@/src/components/LoginModal';

type ChatMessage = {
  id: number;
  user_id: number;
  user_name?: string;
  user_avatar?: string;
  real_name?: string;
  department?: string;
  message_type: 'text' | 'image' | 'file' | 'system';
  content?: string;
  file_name?: string;
  file_url?: string;
  file_size?: number;
  file_ext?: string;
  created_at: string;
};

type DirectMessage = {
  id: number;
  user_id: number;
  sender_type: 'user' | 'admin';
  admin_id: number;
  message_type: 'text' | 'image' | 'file';
  content: string;
  file_name: string;
  file_url: string;
  file_size: number;
  file_ext: string;
  is_read: boolean;
  created_at: string;
  sender_name: string;
  sender_avatar: string;
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
  const { user, isAdmin, loading: authLoading } = useAuth();
  const { showToast } = useToast();
  const [loginOpen, setLoginOpen] = useState(false);
  const [chatMode, setChatMode] = useState<'public' | 'direct'>('public');
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

  // 私聊专用状态
  const [directMessages, setDirectMessages] = useState<DirectMessage[]>([]);
  const [loadingDirect, setLoadingDirect] = useState(false);

  const fileRef = useRef<HTMLInputElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const directBottomRef = useRef<HTMLDivElement>(null);
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
    if (chatMode === 'public') {
      bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
    } else {
      directBottomRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages.length, directMessages.length, chatMode]);

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
      lastDeleteMsRef.current = res.data.deleted_at_ms || 0;
    } catch (err: any) {
      showToast(err.message || '加载消息失败', 'error');
    } finally {
      setLoadingMessages(false);
    }
  };

  const loadDirectMessages = useCallback(async (isSilent = false) => {
    if (!isSilent) setLoadingDirect(true);
    try {
      const res = await api.getDirectMessages({ limit: 100 });
      setDirectMessages((res.data?.messages || []) as DirectMessage[]);
      api.markDirectRead().catch(() => {});
    } catch {
      if (!isSilent) showToast('加载私聊记录失败', 'error');
    } finally {
      if (!isSilent) setLoadingDirect(false);
    }
  }, [showToast]);

  // 当进入聊天并切换模式时，自动加载对应模式的数据并建立同步
  useEffect(() => {
    if (!joined) return;

    if (chatMode === 'public') {
      api.chatJoin()
        .then((res) => {
          if (typeof res.data?.online_count === 'number') {
            setOnlineCount(res.data.online_count);
          }
        })
        .catch(() => {});
      loadHistory();
    } else {
      loadDirectMessages();
      const timer = setInterval(() => {
        loadDirectMessages(true);
      }, 3500);
      return () => clearInterval(timer);
    }
  }, [joined, chatMode, loadDirectMessages]);

  const joinChat = () => {
    if (!user) {
      setLoginOpen(true);
      return;
    }
    setJoining(true);
    setChatMode('public');
    setJoined(true);
    setJoining(false);
  };

  const joinDirectChat = () => {
    if (!user) {
      setLoginOpen(true);
      return;
    }
    setChatMode('direct');
    setJoined(true);
  };

  const leaveChat = async () => {
    try {
      if (chatMode === 'public') {
        await api.chatLeave();
      }
    } catch {}
    setJoined(false);
    setMessages([]);
    setDirectMessages([]);
  };

  useEffect(() => {
    if (!joined || chatMode !== 'public') return;

    const cleanupWS = connectChatWS({
      onMessage: (wsMsg) => {
        if (wsMsg.type === 'message') {
          if (wsMsg.data?.channel === 'direct') {
            loadDirectMessages(true);
          } else {
            mergeMessages([wsMsg.data as ChatMessage]);
          }
        } else if (wsMsg.type === 'presence') {
          if (typeof wsMsg.data?.online_count === 'number') {
            setOnlineCount(wsMsg.data.online_count);
          }
        } else if (wsMsg.type === 'delete') {
          if (wsMsg.data?.ids) {
            applyDeletedMessages(wsMsg.data.ids);
          }
        }
      },
    });

    const pollTimer = setInterval(async () => {
      if (!joinedRef.current) return;
      try {
        const res = await api.getChatMessages({
          afterId: lastMessageIdRef.current,
          afterDeleteMs: lastDeleteMsRef.current,
        });
        if (res.data?.messages?.length) {
          mergeMessages(res.data.messages as ChatMessage[]);
        }
        if (res.data?.deleted_ids?.length) {
          applyDeletedMessages(res.data.deleted_ids);
        }
        if (res.data?.deleted_at_ms) {
          lastDeleteMsRef.current = res.data.deleted_at_ms;
        }
        if (typeof res.data?.online_count === 'number') {
          setOnlineCount(res.data.online_count);
        }
      } catch {}
    }, 4000);

    return () => {
      clearInterval(pollTimer);
      cleanupWS();
    };
  }, [joined, chatMode, loadDirectMessages]);

  const loadEarlier = async () => {
    if (!beforeId || loadingEarlier) return;
    setLoadingEarlier(true);
    try {
      const res = await api.getChatMessages({ beforeId, limit: 30 });
      const earlier = (res.data.messages || []) as ChatMessage[];
      mergeMessages(earlier);
      setHasEarlier(Boolean(res.data.has_more));
      setBeforeId(res.data.next_before_id);
    } catch (err: any) {
      showToast(err.message || '加载历史消息失败', 'error');
    } finally {
      setLoadingEarlier(false);
    }
  };

  const recallMessage = async (msg: ChatMessage) => {
    try {
      await api.deleteChatMessage(msg.id);
      applyDeletedMessages([msg.id]);
      showToast('消息已撤回', 'success');
    } catch (err: any) {
      showToast(err.message || '撤回失败', 'error');
    }
  };

  const sendText = async (e: React.FormEvent) => {
    e.preventDefault();
    const text = content.trim();
    if (!text || sending) return;
    setContent('');
    setSending(true);
    try {
      if (chatMode === 'public') {
        const res = await api.sendChatText(text);
        mergeMessages([res.data as ChatMessage]);
      } else {
        const res = await api.sendDirectText(text);
        if (res.data) {
          setDirectMessages((prev) => [...prev, res.data as DirectMessage]);
        }
      }
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
      if (chatMode === 'public') {
        const res = await api.sendChatFile(fd);
        mergeMessages([res.data as ChatMessage]);
      } else {
        const res = await api.sendDirectFile(fd);
        if (res.data) {
          setDirectMessages((prev) => [...prev, res.data as DirectMessage]);
        }
      }
      showToast('附件发送成功', 'success');
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
        <div className="h-full flex flex-col flex-1 min-h-0 justify-center py-6">
          <div className="sidebar-panel rounded-[2.5rem] p-8 md:p-14 text-center space-y-6 max-w-xl mx-auto w-full shadow-lg border border-border/80">
            <div className="w-20 h-20 rounded-3xl bg-primary/10 mx-auto flex items-center justify-center text-primary shadow-inner">
              <MessageCircle className="w-10 h-10" />
            </div>
            <div className="space-y-2">
              <h2 className="font-serif text-2xl md:text-3xl font-bold text-charcoal">聊天室需要登录</h2>
              <p className="text-sm text-text-muted">请先登录你的会员账号以参与即时交流与文件讨论。</p>
            </div>
            <button onClick={() => setLoginOpen(true)} className="btn-primary mx-auto !py-3 !px-8 flex items-center gap-2">
              <LogIn className="w-4 h-4" />
              立即登录
            </button>
          </div>
        </div>
        <LoginModal open={loginOpen} onClose={() => setLoginOpen(false)} />
      </>
    );
  }

  if (!joined) {
    return (
      <div className="h-full flex flex-col flex-1 min-h-0 justify-center py-6">
        <div className="sidebar-panel rounded-[2.5rem] p-8 md:p-14 text-center space-y-6 max-w-xl mx-auto w-full shadow-lg border border-border/80">
          <div className="w-20 h-20 rounded-3xl bg-primary/10 mx-auto flex items-center justify-center text-primary shadow-inner">
            <Users className="w-10 h-10" />
          </div>
          <div className="space-y-2">
            <h2 className="font-serif text-2xl md:text-3xl font-bold text-charcoal">数学交流聊天室</h2>
            <p className="text-sm text-text-muted max-w-md mx-auto">
              欢迎，{user.nickname || user.username}！加入聊天室与其他会员进行实时交流，或直接与协会管理团队 1 对 1 私信沟通。
            </p>
          </div>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
            <button onClick={joinChat} disabled={joining} className="btn-primary !py-3 !px-6 text-sm flex items-center gap-2 w-full sm:w-auto justify-center shadow-sm cursor-pointer">
              <MessageCircle className="w-4 h-4" />
              {joining ? '正在进入...' : '进入聊天室'}
            </button>
            <button
              onClick={joinDirectChat}
              className="px-5 py-3 rounded-full text-sm font-semibold bg-emerald-50 text-emerald-700 hover:bg-emerald-100/80 border border-emerald-200/80 flex items-center justify-center gap-1.5 w-full sm:w-auto transition-colors shadow-sm cursor-pointer"
            >
              <Headphones className="w-4 h-4 text-emerald-600" />
              与管理员私聊
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col flex-1 min-h-0 pb-1">
      <div className="glass-card rounded-[2rem] overflow-hidden flex flex-col flex-1 min-h-0 border border-border/80 shadow-lg">
        {/* 顶部标题栏 */}
        <div className="px-5 md:px-7 py-3.5 border-b border-border/70 bg-surface/80 backdrop-blur-sm flex items-center justify-between gap-4 shrink-0">
          <div className="flex items-center gap-3">
            <div className={`w-10 h-10 rounded-2xl flex items-center justify-center shrink-0 ${
              chatMode === 'public' ? 'bg-primary/10 text-primary' : 'bg-emerald-50 text-emerald-600'
            }`}>
              {chatMode === 'public' ? <MessageCircle className="w-5 h-5" /> : <Headphones className="w-5 h-5" />}
            </div>
            <div>
              <div className="flex items-center gap-2.5">
                <h2 className="font-serif text-xl font-bold text-charcoal tracking-tight">
                  {chatMode === 'public' ? '公共聊天室' : '与协会管理员私聊'}
                </h2>
                {chatMode === 'public' ? (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700 border border-emerald-200/60">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                    在线 {onlineCount} 人
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700 border border-emerald-200/60">
                    <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
                    官方专属通道
                  </span>
                )}
              </div>
              <p className="text-[11px] text-text-muted mt-0.5">
                {chatMode === 'public' ? '数学爱好者即时交流与学术分享' : '协会管理团队实时为你答疑解惑'}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => setChatMode(chatMode === 'public' ? 'direct' : 'public')}
              className="px-3.5 py-2 rounded-xl text-xs font-semibold bg-canvas-alt hover:bg-black/[0.04] border border-border text-zinc-700 transition-colors flex items-center gap-1.5 cursor-pointer"
            >
              {chatMode === 'public' ? (
                <>
                  <Headphones className="w-3.5 h-3.5 text-emerald-600" />
                  <span>切换管理员私聊</span>
                </>
              ) : (
                <>
                  <MessageCircle className="w-3.5 h-3.5 text-primary" />
                  <span>切换公共大厅</span>
                </>
              )}
            </button>

            <button
              onClick={leaveChat}
              className="btn-secondary !py-2 !px-4 !text-xs text-rose-600 hover:text-rose-700 hover:bg-rose-50 border-rose-200/80 transition-colors flex items-center gap-1.5 shrink-0 rounded-xl cursor-pointer"
              title="离开"
            >
              <X className="w-3.5 h-3.5" />
              <span>退出</span>
            </button>
          </div>
        </div>

        {/* 消息列表主体 */}
        <div className="flex-1 min-h-0 overflow-y-auto bg-canvas-alt/40 px-4 sm:px-6 md:px-8 py-5">
          {chatMode === 'public' ? (
            loadingMessages ? (
              <div className="h-full flex items-center justify-center text-zinc-400 text-sm">
                <div className="flex items-center gap-2">
                  <RotateCcw className="w-4 h-4 animate-spin text-primary" />
                  <span>消息加载中...</span>
                </div>
              </div>
            ) : messages.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-zinc-400 text-sm space-y-2">
                <MessageCircle className="w-10 h-10 text-zinc-300" />
                <span>暂无消息，发条消息打个招呼吧！</span>
              </div>
            ) : (
              <div className="space-y-4 max-w-4xl mx-auto">
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
                    isAdmin={isAdmin}
                    onRecall={() => recallMessage(msg)}
                  />
                ))}
                <div ref={bottomRef} />
              </div>
            )
          ) : (
            // 私聊模式
            loadingDirect ? (
              <div className="h-full flex items-center justify-center text-zinc-400 text-sm">
                <div className="flex items-center gap-2">
                  <RotateCcw className="w-4 h-4 animate-spin text-emerald-600" />
                  <span>私聊记录加载中...</span>
                </div>
              </div>
            ) : directMessages.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-zinc-400 text-sm space-y-2">
                <Headphones className="w-10 h-10 text-emerald-400" />
                <span>暂无私聊消息，在下方输入框向协会管理员留言吧！</span>
              </div>
            ) : (
              <div className="space-y-4 max-w-4xl mx-auto">
                {directMessages.map((msg) => {
                  const isAdmin = msg.sender_type === 'admin';
                  return (
                    <div key={msg.id} className={`flex gap-3 ${!isAdmin ? 'justify-end' : 'justify-start'}`}>
                      {isAdmin && (
                        <div className="w-9 h-9 rounded-full bg-emerald-600 text-white font-bold flex items-center justify-center text-xs shrink-0 mt-0.5 shadow-sm">
                          管
                        </div>
                      )}
                      <div className={`flex flex-col max-w-[78%] ${!isAdmin ? 'items-end' : 'items-start'}`}>
                        <div className={`flex items-center gap-2 text-[11px] text-text-muted mb-1 px-1 ${!isAdmin ? 'justify-end' : ''}`}>
                          <span className="font-medium text-charcoal">{isAdmin ? '协会管理员' : '我'}</span>
                          <span className="inline-flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {formatTime(msg.created_at)}
                          </span>
                        </div>
                        <div className={`w-fit max-w-full rounded-2xl px-4 py-2.5 text-xs sm:text-sm leading-relaxed shadow-sm break-words whitespace-pre-wrap ${
                          !isAdmin
                            ? 'bg-[#4c9eeb] text-white rounded-tr-md selection:bg-white/30 selection:text-white'
                            : 'bg-white border border-border text-charcoal rounded-tl-md selection:bg-accent/10 selection:text-charcoal'
                        }`}>
                          {msg.message_type === 'text' && <div>{msg.content}</div>}
                          {msg.message_type === 'image' && (
                            <ChatFileImage message={{ file_url: msg.file_url, file_name: msg.file_name }} mine={!isAdmin} />
                          )}
                          {msg.message_type === 'file' && (
                            <ChatFileDownload message={{ file_url: msg.file_url, file_name: msg.file_name, file_size: msg.file_size }} mine={!isAdmin} />
                          )}
                        </div>
                      </div>
                      {!isAdmin && (
                        user.avatar ? (
                          <img src={user.avatar} alt="" className="w-9 h-9 rounded-full object-cover border border-border shrink-0 mt-0.5" />
                        ) : (
                          <div className="w-9 h-9 rounded-full bg-pastel-blue text-pastel-blue-text font-bold flex items-center justify-center text-xs shrink-0 mt-0.5">
                            {(user.real_name || user.username).slice(0, 1).toUpperCase()}
                          </div>
                        )
                      )}
                    </div>
                  );
                })}
                <div ref={directBottomRef} />
              </div>
            )
          )}
        </div>

        {/* 底部输入栏：统一高度居中对齐 */}
        <form onSubmit={sendText} className="border-t border-border/80 bg-surface/90 backdrop-blur-sm px-4 py-3 shrink-0">
          <input
            ref={fileRef}
            type="file"
            accept={allowedExts.join(',')}
            onChange={handleFileChange}
            className="hidden"
          />
          <div className="flex items-center gap-2.5 sm:gap-3 max-w-4xl mx-auto">
            <button
              type="button"
              onClick={() => fileRef.current?.click()}
              disabled={uploading}
              className="h-11 w-11 flex items-center justify-center rounded-xl bg-canvas-alt hover:bg-black/[0.04] border border-border shrink-0 transition-colors cursor-pointer"
              title="上传图片或文件 (支持 JPG/PNG/PDF/DOCX 等，最大 5MB)"
              aria-label="上传图片或文件"
            >
              {uploading ? <ImageIcon className="w-4 h-4 animate-pulse text-primary" /> : <Paperclip className="w-4 h-4 text-zinc-600" />}
            </button>
            <input
              type="text"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              maxLength={2000}
              className="app-input h-11 min-w-0 flex-1 rounded-xl px-4 text-sm leading-normal focus:ring-2 focus:ring-primary/20"
              placeholder={chatMode === 'public' ? "输入消息... (按 Enter 发送)" : "向管理员发送私信... (按 Enter 发送)"}
            />
            <button
              type="submit"
              disabled={!content.trim() || sending}
              className="btn-primary h-11 !px-5 shrink-0 rounded-xl shadow-sm flex items-center justify-center gap-2 cursor-pointer"
              title="发送"
              aria-label="发送消息"
            >
              <Send className="w-4 h-4" />
              <span className="hidden sm:inline font-medium text-sm">{sending ? '发送中' : '发送'}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

function MessageItem({
  message,
  mine,
  isAdmin,
  onRecall,
}: {
  message: ChatMessage;
  mine: boolean;
  isAdmin: boolean;
  onRecall: () => void;
}) {
  const [cardOpen, setCardOpen] = useState(false);

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
  const canRecall = isAdmin || (mine && Date.now() - new Date(message.created_at).getTime() <= recallWindowMs);

  return (
    <div className={`flex gap-3 ${mine ? 'justify-end' : 'justify-start'}`}>
      {!mine && (
        <div className="relative mt-5">
          <button
            type="button"
            onClick={() => setCardOpen((v) => !v)}
            className="block h-9 w-9 rounded-full overflow-hidden border border-border cursor-pointer"
            title="查看个人名片"
          >
            {message.user_avatar ? (
              <img src={assetUrl(message.user_avatar)} alt="" className="h-full w-full object-cover" />
            ) : (
              <div className="h-full w-full bg-pastel-blue flex items-center justify-center text-pastel-blue-text text-xs font-bold">
                {avatarText}
              </div>
            )}
          </button>
          {cardOpen && (
            <div className="absolute left-0 top-11 z-10 w-44 rounded-2xl glass-card shadow-xl p-4 space-y-2">
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs text-zinc-400">昵称</span>
                <span className="text-sm font-medium text-charcoal truncate">{message.user_name || `用户 ${message.user_id}`}</span>
              </div>
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs text-zinc-400">姓名</span>
                <span className="text-sm font-medium text-charcoal truncate">{message.real_name || '未填写'}</span>
              </div>
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs text-zinc-400">部门</span>
                <span className="text-sm font-medium text-charcoal truncate">{message.department || '未分配'}</span>
              </div>
            </div>
          )}
        </div>
      )}

      <div className={`flex flex-col max-w-[78%] ${mine ? 'items-end' : 'items-start'}`}>
        <div className={`flex items-center gap-2 text-[11px] text-text-muted mb-1 px-1 ${mine ? 'justify-end' : ''}`}>
          {canRecall && (
            <button
              onClick={onRecall}
              className={`inline-flex items-center gap-1 cursor-pointer transition-colors ${
                isAdmin && !mine ? 'text-rose-500 hover:text-rose-600 font-medium' : 'hover:text-accent'
              }`}
              title={isAdmin && !mine ? '管理员撤回此消息' : '两分钟内可撤回'}
            >
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
        <div className={`w-fit max-w-full rounded-2xl px-4 py-2.5 text-xs sm:text-sm leading-relaxed shadow-sm break-words whitespace-pre-wrap ${
          mine
            ? 'bg-[#4c9eeb] text-white rounded-tr-md selection:bg-white/30 selection:text-white'
            : 'bg-white border border-border text-charcoal rounded-tl-md selection:bg-accent/10 selection:text-charcoal'
        }`}>
          {message.message_type === 'text' && <div>{message.content}</div>}
          {message.message_type === 'image' && (
            <ChatFileImage message={message} mine={mine} />
          )}
          {message.message_type === 'file' && (
            <ChatFileDownload message={message} mine={mine} />
          )}
        </div>
      </div>

      {mine && (
        <div className="relative mt-5 shrink-0">
          <div className="h-9 w-9 rounded-full overflow-hidden border border-border shadow-sm">
            {message.user_avatar ? (
              <img src={assetUrl(message.user_avatar)} alt="" className="h-full w-full object-cover" />
            ) : (
              <div className="h-full w-full bg-pastel-blue flex items-center justify-center text-pastel-blue-text text-xs font-bold">
                {avatarText}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function ChatFileImage({ message, mine }: { message: { file_url?: string; file_name?: string }; mine: boolean }) {
  const [src, setSrc] = useState('');
  const objectUrlRef = useRef('');

  useEffect(() => {
    let cancelled = false;
    if (!message.file_url) return;
    authFetchBlob(message.file_url)
      .then((blob) => {
        if (cancelled) return;
        objectUrlRef.current = URL.createObjectURL(blob);
        setSrc(objectUrlRef.current);
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [message.file_url]);

  useEffect(() => {
    return () => {
      if (objectUrlRef.current) {
        URL.revokeObjectURL(objectUrlRef.current);
        objectUrlRef.current = '';
      }
    };
  }, []);

  if (!src) {
    return (
      <div className="flex h-40 items-center justify-center text-xs text-text-muted">
        图片加载中...
      </div>
    );
  }

  return (
    <a href={src} target="_blank" rel="noreferrer" className="block">
      <img
        src={src}
        alt={message.file_name || '聊天图片'}
        className="max-h-72 max-w-full rounded-xl object-contain"
      />
      <div className={`mt-2 text-xs ${mine ? 'text-white/70' : 'text-text-muted'}`}>
        {message.file_name}
      </div>
    </a>
  );
}

function ChatFileDownload({ message, mine }: { message: { file_url?: string; file_name?: string; file_size?: number }; mine: boolean }) {
  const { showToast } = useToast();
  const [downloading, setDownloading] = useState(false);

  const download = async () => {
    if (!message.file_url || downloading) return;
    setDownloading(true);
    try {
      const blob = await authFetchBlob(message.file_url);
      const objectUrl = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = objectUrl;
      a.download = message.file_name || '文件';
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(objectUrl);
    } catch {
      showToast('文件加载失败，请重试', 'error');
    } finally {
      setDownloading(false);
    }
  };

  return (
    <button
      type="button"
      onClick={download}
      disabled={downloading}
      className={`flex w-full min-w-0 items-center gap-3 rounded-xl p-3 ${
        mine ? 'bg-white/10 hover:bg-white/15' : 'bg-canvas-alt hover:bg-black/[0.03]'
      }`}
    >
      <FileText className="h-8 w-8 shrink-0" />
      <span className="min-w-0 flex-1 text-left">
        <span className="block truncate font-medium">{message.file_name || '文件'}</span>
        <span className={`block text-xs ${mine ? 'text-white/70' : 'text-text-muted'}`}>
          {formatSize(message.file_size)}
        </span>
      </span>
      {downloading ? (
        <RotateCcw className="h-4 w-4 shrink-0 animate-spin" />
      ) : (
        <Download className="h-4 w-4 shrink-0" />
      )}
    </button>
  );
}
