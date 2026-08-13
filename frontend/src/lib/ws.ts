import { tokenStore } from './http';

/**
 * WebSocket 客户端。
 * 鉴权：浏览器 WebSocket 无法自定义 Authorization header，
 * 采用子协议方案 —— 把 token 塞进 Sec-WebSocket-Protocol（bearer.<token>），
 * 不落入 URL 与 access log。
 */
export type ChatWSMessage = {
  type: 'message' | 'presence' | 'delete' | 'notification' | 'pong';
  data: any;
};

type ConnectOptions = {
  onMessage: (msg: ChatWSMessage) => void;
  onClose?: () => void;
  onReconnect?: () => void;
};

const PING_INTERVAL = 30_000;
const RECONNECT_MAX = 30_000;

/**
 * 通用连接：聊天室（/chat/ws）与通知推送（/ws/notify）共用同一套
 * 握手鉴权、心跳与指数退避重连逻辑，仅路径不同。
 */
function connectWS(wsPath: string, opts: ConnectOptions): () => void {
  const WS_BASE = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}${wsPath}`;
  let ws: WebSocket | null = null;
  let disposed = false;
  let retryCount = 0;
  let heartbeatTimer: number | undefined;
  let reconnectTimer: number | undefined;

  const stopHeartbeat = () => {
    if (heartbeatTimer !== undefined) {
      window.clearInterval(heartbeatTimer);
      heartbeatTimer = undefined;
    }
  };

  const startHeartbeat = () => {
    stopHeartbeat();
    heartbeatTimer = window.setInterval(() => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'ping', data: { ts: Date.now() } }));
      }
    }, PING_INTERVAL);
  };

  const connect = () => {
    if (disposed) return;
    const token = tokenStore.getUser();
    if (!token) return;

    ws = new WebSocket(WS_BASE, [`bearer.${token}`]);

    ws.onopen = () => {
      retryCount = 0;
      startHeartbeat();
      opts.onReconnect?.();
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as ChatWSMessage;
        if (msg && typeof msg.type === 'string') opts.onMessage(msg);
      } catch {
        // 忽略无法解析的帧
      }
    };

    ws.onclose = () => {
      stopHeartbeat();
      if (disposed) return;
      opts.onClose?.();
      // 指数退避重连：1s / 2s / 4s ... 封顶 30s
      const delay = Math.min(RECONNECT_MAX, 1000 * 2 ** retryCount);
      retryCount += 1;
      reconnectTimer = window.setTimeout(connect, delay);
    };

    ws.onerror = () => {
      ws?.close();
    };
  };

  connect();

  return () => {
    disposed = true;
    stopHeartbeat();
    if (reconnectTimer !== undefined) window.clearTimeout(reconnectTimer);
    if (ws) {
      ws.onclose = null;
      ws.close();
    }
  };
}

/** 聊天室实时通道（消息/撤回/在线数） */
export function connectChatWS({ onMessage, onClose, onReconnect }: ConnectOptions): () => void {
  return connectWS('/api/v1/chat/ws', { onMessage, onClose, onReconnect });
}

/** 通知实时推送通道：收到 notification 事件即调用 onNotify（红点刷新） */
export function connectNotifyWS(onNotify: () => void): () => void {
  return connectWS('/api/v1/ws/notify', {
    onMessage: (msg) => {
      if (msg.type === 'notification') onNotify();
    },
    onReconnect: onNotify,
  });
}
