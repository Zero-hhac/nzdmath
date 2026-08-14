import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import {
  Search, Send, Paperclip, MessageSquare,
  FileText, Loader2, CheckCircle2,
  RefreshCw
} from 'lucide-react';
import { api } from '@/src/lib/api';
import { useToast } from '@/src/lib/toast';

interface Conversation {
  user_id: number;
  username: string;
  nickname: string;
  real_name: string;
  class_name: string;
  department: string;
  avatar: string;
  last_message: string;
  last_message_type: string;
  last_message_at: string;
  unread_count: number;
}

interface DirectMessage {
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
}

export const DirectChatPanel: React.FC = () => {
  const { showToast } = useToast();
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null);
  const [messages, setMessages] = useState<DirectMessage[]>([]);
  const [keyword, setKeyword] = useState('');
  const [loadingList, setLoadingList] = useState(false);
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [inputText, setInputText] = useState('');
  const [sending, setSending] = useState(false);
  const [uploading, setUploading] = useState(false);

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const selectedUserIdRef = useRef<number | null>(null);

  useEffect(() => {
    selectedUserIdRef.current = selectedUserId;
  }, [selectedUserId]);

  const selectedUser = useMemo(() => {
    if (!selectedUserId) return null;
    return conversations.find((c) => c.user_id === selectedUserId) || null;
  }, [conversations, selectedUserId]);

  const fetchConversations = useCallback(async (isSilent = false) => {
    if (!isSilent) setLoadingList(true);
    try {
      const res = await api.adminListDirectConversations({ keyword: keyword.trim() || undefined, page_size: 50 });
      const list = (res.data?.conversations || []) as Conversation[];
      setConversations(list);
    } catch {
      if (!isSilent) showToast('获取私聊会话列表失败', 'error');
    } finally {
      if (!isSilent) setLoadingList(false);
    }
  }, [keyword, showToast]);

  const fetchMessages = useCallback(async (userId: number, isSilent = false) => {
    if (!isSilent) setLoadingMessages(true);
    try {
      const res = await api.adminGetDirectMessages(userId, { limit: 100 });
      setMessages((res.data?.messages || []) as DirectMessage[]);
      api.adminMarkDirectRead(userId).catch(() => {});
    } catch {
      if (!isSilent) showToast('加载对话记录失败', 'error');
    } finally {
      if (!isSilent) setLoadingMessages(false);
    }
  }, [showToast]);

  useEffect(() => {
    fetchConversations();
  }, [fetchConversations]);

  useEffect(() => {
    const timer = setInterval(() => {
      fetchConversations(true);
      if (selectedUserIdRef.current) {
        fetchMessages(selectedUserIdRef.current, true);
      }
    }, 3500);
    return () => clearInterval(timer);
  }, [fetchConversations, fetchMessages]);

  useEffect(() => {
    if (selectedUserId) {
      fetchMessages(selectedUserId, false);
    } else {
      setMessages([]);
    }
  }, [selectedUserId, fetchMessages]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSendMessage = async () => {
    if (!selectedUserId || !inputText.trim() || sending) return;
    const text = inputText.trim();
    setInputText('');
    setSending(true);
    try {
      const res = await api.adminSendDirectText(selectedUserId, text);
      if (res.data) {
        setMessages((prev) => [...prev, res.data as DirectMessage]);
      }
      fetchConversations(true);
    } catch (e: any) {
      showToast(e?.message || '发送失败', 'error');
    } finally {
      setSending(false);
    }
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !selectedUserId || uploading) return;
    if (file.size > 5 * 1024 * 1024) {
      showToast('文件大小不能超过 5MB', 'error');
      return;
    }
    const formData = new FormData();
    formData.append('file', file);
    setUploading(true);
    try {
      const res = await api.adminSendDirectFile(selectedUserId, formData);
      if (res.data) {
        setMessages((prev) => [...prev, res.data as DirectMessage]);
      }
      showToast('附件已发送', 'success');
      fetchConversations(true);
    } catch (e: any) {
      showToast(e?.message || '上传发送失败', 'error');
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  };

  return (
    <div className="bg-white rounded-[2rem] border border-border shadow-sm overflow-hidden flex flex-col md:flex-row h-[calc(100vh-210px)] min-h-[500px]">
      {/* 左侧：会话列表 */}
      <div className="w-full md:w-80 lg:w-96 border-r border-border flex flex-col bg-canvas-alt/50 shrink-0">
        <div className="p-4 border-b border-border space-y-3 bg-white/70 backdrop-blur">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <MessageSquare className="w-5 h-5 text-primary" />
              <span className="font-bold text-charcoal">会员私信列表</span>
            </div>
            <button
              onClick={() => fetchConversations()}
              className="p-1.5 rounded-full hover:bg-zinc-100 text-zinc-500 transition-colors cursor-pointer"
              title="刷新列表"
            >
              <RefreshCw className={`w-4 h-4 ${loadingList ? 'animate-spin' : ''}`} />
            </button>
          </div>
          <div className="relative">
            <Search className="w-4 h-4 text-zinc-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && fetchConversations()}
              placeholder="搜索姓名 / 昵称 / 班级..."
              className="w-full pl-9 pr-3 py-2 bg-zinc-100 rounded-xl text-xs text-charcoal placeholder-zinc-400 focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/20 border border-transparent focus:border-primary/30 transition-all"
            />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto divide-y divide-border/60">
          {loadingList && conversations.length === 0 ? (
            <div className="py-20 flex flex-col items-center justify-center text-zinc-400 gap-2">
              <Loader2 className="w-6 h-6 animate-spin text-primary" />
              <span className="text-xs">加载会话中...</span>
            </div>
          ) : conversations.length === 0 ? (
            <div className="py-20 text-center text-zinc-400 text-xs px-4">
              暂无会员私聊记录
            </div>
          ) : (
            conversations.map((conv) => {
              const isSelected = selectedUserId === conv.user_id;
              const displayName = conv.real_name || conv.nickname || conv.username || `会员 #${conv.user_id}`;
              return (
                <button
                  key={conv.user_id}
                  onClick={() => setSelectedUserId(conv.user_id)}
                  className={`w-full text-left p-3.5 flex items-start gap-3 transition-colors cursor-pointer ${
                    isSelected ? 'bg-primary/10 border-l-4 border-primary' : 'hover:bg-zinc-100/70'
                  }`}
                >
                  <div className="relative shrink-0">
                    {conv.avatar ? (
                      <img src={conv.avatar} alt="" className="w-10 h-10 rounded-full object-cover border border-border" />
                    ) : (
                      <div className="w-10 h-10 rounded-full bg-pastel-blue text-pastel-blue-text font-bold flex items-center justify-center text-sm border border-border">
                        {displayName.slice(0, 1).toUpperCase()}
                      </div>
                    )}
                    {conv.unread_count > 0 && (
                      <span className="absolute -top-1 -right-1 bg-rose-500 text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full min-w-[18px] text-center shadow-sm animate-pulse">
                        {conv.unread_count > 99 ? '99+' : conv.unread_count}
                      </span>
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-1 mb-0.5">
                      <span className="font-semibold text-xs text-charcoal truncate">{displayName}</span>
                      <span className="text-[10px] text-zinc-400 shrink-0">
                        {conv.last_message_at ? new Date(conv.last_message_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : ''}
                      </span>
                    </div>
                    <div className="text-[11px] text-zinc-500 flex items-center gap-1.5 mb-1">
                      {conv.department && <span className="bg-zinc-200/70 text-zinc-700 px-1.5 py-0.2 rounded text-[10px]">{conv.department}</span>}
                      {conv.class_name && <span className="text-zinc-400 text-[10px] truncate">{conv.class_name}</span>}
                    </div>
                    <p className="text-xs text-zinc-600 truncate font-normal">
                      {conv.last_message_type === 'image' ? '[图片]' : conv.last_message_type === 'file' ? '[文件]' : conv.last_message || '无内容'}
                    </p>
                  </div>
                </button>
              );
            })
          )}
        </div>
      </div>

      {/* 右侧：聊天主视窗 */}
      <div className="flex-1 flex flex-col bg-white">
        {selectedUser ? (
          <>
            {/* 顶栏用户名片 */}
            <div className="px-6 py-3.5 border-b border-border flex items-center justify-between bg-canvas-alt/30 shrink-0">
              <div className="flex items-center gap-3">
                {selectedUser.avatar ? (
                  <img src={selectedUser.avatar} alt="" className="w-10 h-10 rounded-full object-cover border border-border" />
                ) : (
                  <div className="w-10 h-10 rounded-full bg-pastel-blue text-pastel-blue-text font-bold flex items-center justify-center border border-border">
                    {(selectedUser.real_name || selectedUser.username).slice(0, 1).toUpperCase()}
                  </div>
                )}
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="font-bold text-charcoal text-sm">
                      {selectedUser.real_name ? `${selectedUser.real_name} (${selectedUser.username})` : selectedUser.username}
                    </h3>
                    {selectedUser.department && (
                      <span className="text-[10px] font-semibold bg-primary/10 text-primary px-2 py-0.5 rounded-full">
                        {selectedUser.department}
                      </span>
                    )}
                  </div>
                  <p className="text-[11px] text-zinc-400">
                    {selectedUser.class_name || '班级未填'} · 用户ID: #{selectedUser.user_id}
                  </p>
                </div>
              </div>
              <div className="text-xs text-zinc-400 flex items-center gap-1.5">
                <CheckCircle2 className="w-4 h-4 text-emerald-500" />
                <span>与管理员私聊通道</span>
              </div>
            </div>

            {/* 消息历史滚动区 */}
            <div className="flex-1 overflow-y-auto p-6 space-y-4 bg-zinc-50/40">
              {loadingMessages ? (
                <div className="h-full flex flex-col items-center justify-center text-zinc-400 gap-2">
                  <Loader2 className="w-6 h-6 animate-spin text-primary" />
                  <span className="text-xs">加载记录中...</span>
                </div>
              ) : messages.length === 0 ? (
                <div className="h-full flex items-center justify-center text-zinc-400 text-xs">
                  暂无消息，你可以直接在下方输入框向该会员发送第一条信息。
                </div>
              ) : (
                messages.map((msg) => {
                  const isAdmin = msg.sender_type === 'admin';
                  return (
                    <div key={msg.id} className={`flex gap-3 ${isAdmin ? 'justify-end' : 'justify-start'}`}>
                      {!isAdmin && (
                        selectedUser.avatar ? (
                          <img src={selectedUser.avatar} alt="" className="w-9 h-9 rounded-full object-cover border border-border shrink-0 mt-0.5" />
                        ) : (
                          <div className="w-9 h-9 rounded-full bg-pastel-blue text-pastel-blue-text font-bold flex items-center justify-center text-xs shrink-0 mt-0.5">
                            {(selectedUser.real_name || selectedUser.username).slice(0, 1).toUpperCase()}
                          </div>
                        )
                      )}

                      {/* 仿 QQ 气泡流式布局：w-fit 自适应文字长度，到达 max-w-[75%] 自动折行 */}
                      <div className={`flex flex-col max-w-[75%] ${isAdmin ? 'items-end' : 'items-start'}`}>
                        <div className={`flex items-center gap-2 text-[11px] text-zinc-400 mb-1 px-1 ${isAdmin ? 'justify-end' : ''}`}>
                          <span className="font-medium text-zinc-600">{isAdmin ? '管理员（你）' : (msg.sender_name || selectedUser.real_name || selectedUser.username)}</span>
                          <span className="text-[10px] text-zinc-400">
                            {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                          </span>
                        </div>

                        <div
                          className={`w-fit max-w-full px-4 py-2.5 rounded-2xl text-xs sm:text-sm leading-relaxed shadow-sm break-words whitespace-pre-wrap ${
                            isAdmin
                              ? 'bg-[#4c9eeb] text-white rounded-tr-sm selection:bg-white/30 selection:text-white'
                              : 'bg-white border border-border text-charcoal rounded-tl-sm selection:bg-accent/10 selection:text-charcoal'
                          }`}
                        >
                          {msg.message_type === 'image' ? (
                            <a href={msg.file_url} target="_blank" rel="noreferrer" className="block overflow-hidden rounded-xl">
                              <img src={msg.file_url} alt={msg.file_name} className="max-h-60 max-w-full rounded-xl object-cover hover:scale-105 transition-transform" />
                            </a>
                          ) : msg.message_type === 'file' ? (
                            <a
                              href={msg.file_url}
                              target="_blank"
                              rel="noreferrer"
                              className={`flex items-center gap-2 font-medium underline underline-offset-2 ${
                                isAdmin ? 'text-white hover:text-white/80' : 'text-primary hover:text-primary-hover'
                              }`}
                            >
                              <FileText className="w-4 h-4 shrink-0" />
                              <span className="truncate">{msg.file_name || '附件'}</span>
                              <span className="text-[10px] opacity-75 shrink-0">
                                ({msg.file_size ? `${Math.round(msg.file_size / 1024)}KB` : ''})
                              </span>
                            </a>
                          ) : (
                            <span>{msg.content}</span>
                          )}
                        </div>
                      </div>

                      {isAdmin && (
                        <div className="w-9 h-9 rounded-full bg-[#4c9eeb] text-white font-bold flex items-center justify-center text-xs shrink-0 mt-0.5 shadow-sm">
                          管
                        </div>
                      )}
                    </div>
                  );
                })
              )}
              <div ref={messagesEndRef} />
            </div>

            {/* 底部输入框 */}
            <div className="p-4 border-t border-border bg-white flex items-center gap-2 shrink-0">
              <input
                type="file"
                ref={fileInputRef}
                onChange={handleFileUpload}
                className="hidden"
              />
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={uploading}
                className="w-10 h-10 rounded-xl bg-zinc-100 hover:bg-zinc-200 text-zinc-600 flex items-center justify-center transition-colors shrink-0 disabled:opacity-50 cursor-pointer"
                title="发送图片或附件 (≤5MB)"
              >
                {uploading ? <Loader2 className="w-4 h-4 animate-spin text-primary" /> : <Paperclip className="w-4 h-4" />}
              </button>
              <div className="flex-1 relative flex items-center">
                <input
                  type="text"
                  value={inputText}
                  onChange={(e) => setInputText(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && !e.shiftKey) {
                      e.preventDefault();
                      handleSendMessage();
                    }
                  }}
                  placeholder={`回复给 ${selectedUser.real_name || selectedUser.username}...`}
                  className="w-full h-10 px-4 bg-zinc-100 rounded-xl text-xs text-charcoal placeholder-zinc-400 focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/20 border border-transparent focus:border-primary/30 transition-all"
                />
              </div>
              <button
                type="button"
                onClick={handleSendMessage}
                disabled={sending || !inputText.trim()}
                className="h-10 px-5 rounded-xl bg-primary hover:bg-primary-hover text-white text-xs font-semibold flex items-center gap-1.5 transition-colors shrink-0 disabled:opacity-50 disabled:cursor-not-allowed shadow-sm cursor-pointer"
              >
                {sending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
                <span>发送</span>
              </button>
            </div>
          </>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-zinc-400 p-8 text-center">
            <div className="w-16 h-16 rounded-full bg-zinc-100 flex items-center justify-center mb-3">
              <MessageSquare className="w-8 h-8 text-zinc-300" />
            </div>
            <h4 className="font-bold text-zinc-600 text-sm mb-1">选择一个会员以开始沟通</h4>
            <p className="text-xs text-zinc-400 max-w-xs">
              从左侧列表中选择任意会员，可查看完整沟通记录并进行实时 1 对 1 回复与文件传递。
            </p>
          </div>
        )}
      </div>
    </div>
  );
};
